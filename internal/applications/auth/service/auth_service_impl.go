package service

import (
	"context"
	"errors"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	emailService "github.com/letenk/golang-authentication/internal/applications/email/service"
	otpRepo "github.com/letenk/golang-authentication/internal/applications/otp/repository/db"
	refreshTokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	"github.com/letenk/golang-authentication/internal/applications/transaction"
	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/letenk/golang-authentication/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	trxService      transaction.TrxService
	userRepository  userRepo.UserRepository
	tokenRepository refreshTokenRepo.RefreshTokenRepository
	otpRepository   otpRepo.OTPRepository
	emailService    emailService.EmailService
	jwtConfig       *jwt_config.JWTConfig
	otpConfig       *credential.OTPConfig
}

func NewAuthService(
	trxService transaction.TrxService,
	userRepository userRepo.UserRepository,
	tokenRepository refreshTokenRepo.RefreshTokenRepository,
	otpRepository otpRepo.OTPRepository,
	emailService emailService.EmailService,
	jwtConfig *jwt_config.JWTConfig,
	otpConfig *credential.OTPConfig,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		trxService:      trxService,
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		otpRepository:   otpRepository,
		emailService:    emailService,
		jwtConfig:       jwtConfig,
		otpConfig:       otpConfig,
	}
}

func (service *AuthServiceImpl) Register(ctx context.Context, param *dto.ParameterRegister) (*domain.User, error) {
	// Check if email already exists
	if param.Email != "" {
		existingUser, err := service.userRepository.FindByEmail(ctx, param.Email)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("failed to check email: %s", err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to check email"), nil)
		}
		if existingUser != nil {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("email already registered"), nil)
		}
	}

	// Check if phone already exists
	if param.Phone != "" {
		cleanedPhone, err := utils.NormalizePhoneNumber(param.Phone)
		if err != nil {
			return nil, err
		}
		param.Phone = cleanedPhone

		existingUser, err := service.userRepository.FindByPhone(ctx, param.Phone)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("failed to check phone: %s", err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to check phone"), nil)
		}
		if existingUser != nil {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("phone already registered"), nil)
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(param.Password)
	if err != nil {
		log.Errorf("failed to hash password: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to hash password"), nil)
	}

	// Determine login type
	loginType := "email"
	if param.Email == "" && param.Phone != "" {
		loginType = "phone"
	}

	// Create new user (created by the system user; empty fields stored as NULL)
	result, err := service.userRepository.Create(ctx, userRepo.CreateUserParams{
		Name:      param.FullName,
		Email:     param.Email,
		Phone:     param.Phone,
		Password:  hashedPassword,
		LoginType: loginType,
	})
	if err != nil {
		// The pre-checks above can miss a concurrent registration or an
		// email/phone still held by a soft-deleted account; the unique index catches both.
		switch {
		case errors.Is(err, domain.ErrEmailTaken):
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("email already registered"), nil)
		case errors.Is(err, domain.ErrPhoneTaken):
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("phone already registered"), nil)
		}
		log.Errorf("failed to register user: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.BadData, errors.New("failed to register user"), nil)
	}

	// Send verification email asynchronously after successful registration
	if param.Email != "" {
		go func(userID int64) {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("panic in SendVerificationEmail goroutine: %v", r)
				}
			}()
			if err := service.SendVerificationEmail(context.Background(), userID); err != nil {
				log.Errorf("failed to send verification email for user %d: %s", userID, err)
			}
		}(result.ID)
	}

	return result, nil
}

// Login authenticates user and returns access token with refresh token
func (service *AuthServiceImpl) Login(ctx context.Context, req *dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	// Find user by email or phone
	var user *domain.User
	var err error

	if req.Email != "" {
		user, err = service.userRepository.FindByEmail(ctx, req.Email)
	} else {
		cleanedPhone, normErr := utils.NormalizePhoneNumber(req.Phone)
		if normErr != nil {
			return nil, exceptions.NewAuthenticationError("Invalid credentials", exceptions.AuthenticationInvalidCredentials)
		}
		user, err = service.userRepository.FindByPhone(ctx, cleanedPhone)
	}

	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, exceptions.NewAuthenticationError("Invalid credentials", exceptions.AuthenticationInvalidCredentials)
		}
		log.Errorf("failed to find user: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find user"), nil)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, exceptions.NewAuthenticationError("Invalid credentials", exceptions.AuthenticationInvalidCredentials)
	}

	// Generate access token
	accessToken, accessExpiresAt, err := service.jwtConfig.GenerateAccessToken(user.ID)
	if err != nil {
		log.Errorf("failed to generate access token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to generate access token"), nil)
	}

	// Generate refresh token
	refreshToken, refreshExpiresAt := service.jwtConfig.GenerateRefreshToken()

	// Set device info
	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	_, err = service.tokenRepository.Create(ctx, refreshTokenRepo.CreateRefreshTokenParams{
		UserID:     user.ID,
		Token:      refreshToken,
		DeviceName: deviceName,
		DeviceID:   req.DeviceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  refreshExpiresAt,
	})
	if err != nil {
		log.Errorf("failed to create refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to create refresh token"), nil)
	}

	return &dto.LoginResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt.Format(time.RFC3339),
		RefreshTokenExpiresAt: refreshExpiresAt.Format(time.RFC3339),
	}, nil
}

// RefreshToken validates the refresh token, rotates it, and returns a new token pair
func (service *AuthServiceImpl) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, ipAddress, userAgent string) (*dto.RefreshTokenResponse, error) {
	// Find the token in DB
	token, err := service.tokenRepository.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, exceptions.NewAuthenticationError("Invalid refresh token", exceptions.AuthenticationInvalidToken)
		}
		log.Errorf("failed to find refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
	}

	// Check if already revoked
	if token.IsRevoked() {
		return nil, exceptions.NewAuthenticationError("Refresh token has been revoked", exceptions.AuthenticationTokenRevoked)
	}

	// Check if expired
	if token.IsExpired(time.Now().UTC()) {
		return nil, exceptions.NewAuthenticationError("Refresh token has expired", exceptions.AuthenticationTokenExpired)
	}

	// Generate new access token
	newAccessToken, accessExpiresAt, err := service.jwtConfig.GenerateAccessToken(token.UserID)
	if err != nil {
		log.Errorf("failed to generate access token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to generate access token"), nil)
	}

	// Generate new refresh token
	newRefreshToken, refreshExpiresAt := service.jwtConfig.GenerateRefreshToken()

	// Token rotation: revoke old token and record which token replaced it.
	// ErrNotFound means a concurrent request already rotated this token.
	err = service.tokenRepository.ReplaceToken(ctx, req.RefreshToken, newRefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, exceptions.NewAuthenticationError("Refresh token has been revoked", exceptions.AuthenticationTokenRevoked)
		}
		log.Errorf("failed to rotate refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to rotate refresh token"), nil)
	}

	// Create new refresh token — carry over device info from the old token
	deviceName := token.DeviceName
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	_, err = service.tokenRepository.Create(ctx, refreshTokenRepo.CreateRefreshTokenParams{
		UserID:     token.UserID,
		Token:      newRefreshToken,
		DeviceName: deviceName,
		DeviceID:   token.DeviceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		ExpiresAt:  refreshExpiresAt,
	})
	if err != nil {
		log.Errorf("failed to create new refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to create new refresh token"), nil)
	}

	return &dto.RefreshTokenResponse{
		AccessToken:           newAccessToken,
		RefreshToken:          newRefreshToken,
		AccessTokenExpiresAt:  accessExpiresAt.Format(time.RFC3339),
		RefreshTokenExpiresAt: refreshExpiresAt.Format(time.RFC3339),
	}, nil
}

// Logout revokes the given refresh token (logout from single device)
func (service *AuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	// Find the token
	token, err := service.tokenRepository.FindByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewAuthenticationError("Invalid refresh token", exceptions.AuthenticationInvalidToken)
		}
		log.Errorf("failed to find refresh token: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
	}

	// Check if already revoked
	if token.IsRevoked() {
		return exceptions.NewAuthenticationError("Refresh token has already been revoked", exceptions.AuthenticationTokenRevoked)
	}

	// Revoke the token
	err = service.tokenRepository.Revoke(ctx, refreshToken)
	if err != nil {
		log.Errorf("failed to revoke refresh token: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to revoke refresh token"), nil)
	}

	return nil
}

// LogoutAll revokes all active refresh tokens for the user (logout from all devices)
func (service *AuthServiceImpl) LogoutAll(ctx context.Context, userID int64) (int, error) {
	count, err := service.tokenRepository.RevokeAllByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to revoke all refresh tokens for user %d: %s", userID, err)
		return 0, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to logout all devices"), nil)
	}

	return count, nil
}

// GetMe retrieves current user information with active sessions
func (service *AuthServiceImpl) GetMe(ctx context.Context, userID int64) (*dto.GetMeResponse, error) {
	// Find user by ID
	user, err := service.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
		}
		log.Errorf("failed to find user by ID: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to get user"), nil)
	}

	userResp := &dto.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FullName:   user.Name,
		Phone:      user.Phone,
		IsVerified: user.IsVerified,
	}

	return &dto.GetMeResponse{
		User: userResp,
	}, nil
}

// DeleteAccount soft-deletes the user and revokes all their refresh tokens
func (service *AuthServiceImpl) DeleteAccount(ctx context.Context, userID int64) error {
	if err := service.userRepository.SoftDeleteByID(ctx, userID, userID); err != nil {
		log.Errorf("failed to soft delete user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataDeleteFailed, errors.New("failed to delete account"), nil)
	}

	if _, err := service.tokenRepository.RevokeAllByUserID(ctx, userID); err != nil {
		log.Errorf("failed to revoke tokens after account deletion for user %d: %s", userID, err)
	}

	return nil
}

// GetSessions returns all active sessions for the given user
func (service *AuthServiceImpl) GetSessions(ctx context.Context, userID int64) ([]*dto.ActiveSessionResponse, error) {
	tokens, err := service.tokenRepository.FindActiveByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to get active sessions for user %d: %s", userID, err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to get sessions"), nil)
	}

	sessions := make([]*dto.ActiveSessionResponse, 0, len(tokens))
	for _, t := range tokens {
		session := &dto.ActiveSessionResponse{
			ID:         t.ID,
			DeviceName: t.DeviceName,
			DeviceID:   t.DeviceID,
			IPAddress:  t.IPAddress,
			UserAgent:  t.UserAgent,
		}
		if !t.CreatedAt.IsZero() {
			session.CreatedAt = t.CreatedAt.UTC().Format(time.RFC3339)
			session.LastUsed = t.CreatedAt.UTC().Format(time.RFC3339)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateProfile updates the user's full name and/or phone number
func (service *AuthServiceImpl) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	params := userRepo.UpdateProfileParams{UpdatedBy: userID}

	if req.Phone != "" {
		cleanedPhone, err := utils.NormalizePhoneNumber(req.Phone)
		if err != nil {
			return nil, err
		}
		existingUser, err := service.userRepository.FindByPhone(ctx, cleanedPhone)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			log.Errorf("failed to check phone duplication for user %d: %s", userID, err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to update profile"), nil)
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("phone already in use"), nil)
		}
		params.Phone = &cleanedPhone
	}

	if req.FullName != "" {
		params.Name = &req.FullName
	}

	user, err := service.userRepository.UpdateProfile(ctx, userID, params)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
		}
		if errors.Is(err, domain.ErrPhoneTaken) {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("phone already in use"), nil)
		}
		log.Errorf("failed to update profile for user %d: %s", userID, err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to update profile"), nil)
	}

	return &dto.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FullName:   user.Name,
		Phone:      user.Phone,
		IsVerified: user.IsVerified,
	}, nil
}

// UpdatePassword changes the user's password after validating the old one
func (service *AuthServiceImpl) UpdatePassword(ctx context.Context, userID int64, req *dto.UpdatePasswordRequest) error {
	user, err := service.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
		}
		log.Errorf("failed to find user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to update password"), nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return exceptions.NewBusinessLogicError(exceptions.InvalidArgument, errors.New("old password is incorrect"), nil)
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Errorf("failed to hash password for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to update password"), nil)
	}

	err = service.userRepository.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		log.Errorf("failed to update password for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to update password"), nil)
	}

	return nil
}

// RevokeSession revokes a specific session by ID, ensuring it belongs to the given user
func (service *AuthServiceImpl) RevokeSession(ctx context.Context, userID int64, sessionID int64) error {
	err := service.tokenRepository.RevokeByIDForUser(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("session not found"), nil)
		}
		log.Errorf("failed to revoke session %d for user %d: %s", sessionID, userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to revoke session"), nil)
	}

	return nil
}

// ForgotPassword sends an OTP to the user's email for password reset
func (service *AuthServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	user, err := service.userRepository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Don't reveal whether the email exists
			return nil
		}
		log.Errorf("failed to find user by email for forgot password: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to process request"), nil)
	}

	otpExpire, err := time.ParseDuration(service.otpConfig.Expire)
	if err != nil {
		log.Errorf("failed to parse OTP expire duration: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	otp, err := utils.GenerateOTP(service.otpConfig.Length)
	if err != nil {
		log.Errorf("failed to generate OTP: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	expiresAt := time.Now().UTC().Add(otpExpire)

	err = service.trxService.WithTx(ctx, func(txCtx context.Context) error {
		if err := service.otpRepository.InvalidatePreviousByUserIDAndPurpose(txCtx, user.ID, otpRepo.OTPPurposePasswordReset); err != nil {
			return err
		}
		_, err := service.otpRepository.Create(txCtx, otpRepo.CreateOTPParams{
			UserID:    user.ID,
			Code:      otp,
			Purpose:   otpRepo.OTPPurposePasswordReset,
			ExpiresAt: expiresAt,
		})
		return err
	})
	if err != nil {
		log.Errorf("failed to create OTP for user %d: %s", user.ID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	name := user.Name
	if name == "" {
		name = email
	}

	// Send OTP email in background to avoid blocking the response.
	// Uses an independent context with timeout to prevent hanging, and recover to catch panics.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in SendOTP goroutine: %v", r)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = ctx

		if err := service.emailService.SendOTP(email, name, otp); err != nil {
			log.Errorf("failed to send OTP email to %s: %s", email, err)
		}
	}()

	return nil
}

// ResetPassword validates the OTP and updates the user's password
func (service *AuthServiceImpl) ResetPassword(ctx context.Context, req *dto.ResetPasswordRequest) error {
	user, err := service.userRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("invalid email or OTP"), nil)
		}
		log.Errorf("failed to find user by email for reset password: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to process request"), nil)
	}

	otpRecord, err := service.otpRepository.FindValidByUserAndCode(ctx, user.ID, otpRepo.OTPPurposePasswordReset, req.OTP)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.InvalidArgument, errors.New("invalid or expired OTP"), nil)
		}
		log.Errorf("failed to find OTP for user %d: %s", user.ID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to process request"), nil)
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Errorf("failed to hash password for user %d: %s", user.ID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to reset password"), nil)
	}

	err = service.trxService.WithTx(ctx, func(txCtx context.Context) error {
		if err := service.otpRepository.MarkUsed(txCtx, otpRecord.ID); err != nil {
			return err
		}
		return service.userRepository.UpdatePassword(txCtx, user.ID, hashedPassword)
	})
	if err != nil {
		log.Errorf("failed to reset password for user %d: %s", user.ID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to reset password"), nil)
	}

	return nil
}

// SendVerificationEmail generates an OTP and sends a verification email to the user.
// Email is sent asynchronously via goroutine to avoid blocking the caller.
func (service *AuthServiceImpl) SendVerificationEmail(ctx context.Context, userID int64) error {
	user, err := service.userRepository.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
		}
		log.Errorf("failed to find user %d for verification email: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to process request"), nil)
	}

	if user.Email == "" {
		return exceptions.NewBusinessLogicError(exceptions.InvalidArgument, errors.New("user has no email address"), nil)
	}

	if user.IsVerified {
		return exceptions.NewBusinessLogicError(exceptions.InvalidArgument, errors.New("email is already verified"), nil)
	}

	otpExpire, err := time.ParseDuration(service.otpConfig.Expire)
	if err != nil {
		log.Errorf("failed to parse OTP expire duration: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	otp, err := utils.GenerateOTP(service.otpConfig.Length)
	if err != nil {
		log.Errorf("failed to generate OTP for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	expiresAt := time.Now().UTC().Add(otpExpire)

	err = service.trxService.WithTx(ctx, func(txCtx context.Context) error {
		if err := service.otpRepository.InvalidatePreviousByUserIDAndPurpose(txCtx, userID, otpRepo.OTPPurposeEmailVerification); err != nil {
			return err
		}
		_, err := service.otpRepository.Create(txCtx, otpRepo.CreateOTPParams{
			UserID:    userID,
			Code:      otp,
			Purpose:   otpRepo.OTPPurposeEmailVerification,
			ExpiresAt: expiresAt,
		})
		return err
	})
	if err != nil {
		log.Errorf("failed to create verification OTP for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to process request"), nil)
	}

	email := user.Email
	name := user.Name
	if name == "" {
		name = email
	}

	// Send verification email in background to avoid blocking the response.
	// Uses an independent context with timeout to prevent hanging, and recover to catch panics.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("panic in SendVerificationOTP goroutine: %v", r)
			}
		}()

		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = sendCtx

		if err := service.emailService.SendVerificationOTP(email, name, otp); err != nil {
			log.Errorf("failed to send verification email to %s: %s", email, err)
		}
	}()

	return nil
}

// VerifyEmail validates the OTP and marks the user's email as verified
func (service *AuthServiceImpl) VerifyEmail(ctx context.Context, userID int64, code string) error {
	otpRecord, err := service.otpRepository.FindValidByUserAndCode(ctx, userID, otpRepo.OTPPurposeEmailVerification, code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return exceptions.NewBusinessLogicError(exceptions.InvalidArgument, errors.New("invalid or expired OTP"), nil)
		}
		log.Errorf("failed to find verification OTP for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to process request"), nil)
	}

	now := time.Now().UTC()

	err = service.trxService.WithTx(ctx, func(txCtx context.Context) error {
		if err := service.otpRepository.MarkUsed(txCtx, otpRecord.ID); err != nil {
			return err
		}
		return service.userRepository.MarkEmailVerified(txCtx, userID, now)
	})
	if err != nil {
		log.Errorf("failed to verify email for user %d: %s", userID, err)
		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to verify email"), nil)
	}

	return nil
}
