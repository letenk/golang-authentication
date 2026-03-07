package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	refreshTokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
	"github.com/letenk/golang-authentication/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	userRepository  userRepo.UserRepository
	tokenRepository refreshTokenRepo.RefreshTokenRepository
	jwtConfig       *jwt_config.JWTConfig
}

func NewAuthService(
	userRepository userRepo.UserRepository,
	tokenRepository refreshTokenRepo.RefreshTokenRepository,
	jwtConfig *jwt_config.JWTConfig,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
		jwtConfig:       jwtConfig,
	}
}

func (service *AuthServiceImpl) Register(ctx context.Context, param *dto.ParameterRegister) (*models.User, error) {
	// Check if email already exists
	if param.Email != "" {
		existingUser, err := service.userRepository.FindByEmail(ctx, param.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Errorf("failed to check email: %s", err)
			return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to check email"), nil)
		}
		if existingUser != nil {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("email already registered"), nil)
		}
	}

	// Check if phone already exists
	if param.Phone != "" {
		cleanedPhone := utils.EnsurePhoneNumber(param.Phone)
		param.Phone = cleanedPhone

		existingUser, err := service.userRepository.FindByPhone(ctx, param.Phone)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
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

	// Prepare user data
	data := models.UserSetter{
		Password:  omit.From(hashedPassword),
		CreatedBy: omit.From(int64(0)),
		UpdatedBy: omit.From(int64(0)),
	}

	// Set optional fields
	if param.FullName != "" {
		data.Name = omitnull.From(param.FullName)
	}

	if param.Phone != "" {
		data.Phone = omitnull.From(param.Phone)
	}

	if param.Email != "" {
		data.Email = omitnull.From(param.Email)
	}

	// Determine login type
	loginType := "email"
	if param.Email == "" && param.Phone != "" {
		loginType = "phone"
	}
	data.LoginType = omit.From(loginType)

	// Set default values
	data.IsVerified = omitnull.From(false)

	// Create new user
	result, err := service.userRepository.Create(ctx, &data)
	if err != nil {
		log.Errorf("failed to register user: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.BadData, errors.New("failed to register user"), nil)
	}

	return result, nil
}

// Login authenticates user and returns access token with refresh token
func (service *AuthServiceImpl) Login(ctx context.Context, req *dto.LoginRequest, ipAddress, userAgent string) (*dto.LoginResponse, error) {
	// Find user by email
	user, err := service.userRepository.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewAuthenticationError("Invalid email or password", exceptions.AuthenticationInvalidCredentials)
		}
		log.Errorf("failed to find user by email: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find user"), nil)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, exceptions.NewAuthenticationError("Invalid email or password", exceptions.AuthenticationInvalidCredentials)
	}

	// Generate access token
	email := ""
	if user.Email.IsValue() {
		email = user.Email.GetOrZero()
	}

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

	// Create refresh token record using Setter
	tokenSetter := &models.RefreshTokenSetter{
		UserID:     omit.From(user.ID),
		Token:      omit.From(refreshToken),
		DeviceName: omitnull.From(deviceName),
		DeviceID:   omitnull.From(req.DeviceID),
		IPAddress:  omitnull.From(ipAddress),
		UserAgent:  omitnull.From(userAgent),
		ExpiresAt:  omit.From(refreshExpiresAt),
	}

	_, err = service.tokenRepository.Create(ctx, tokenSetter)
	if err != nil {
		log.Errorf("failed to create refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to create refresh token"), nil)
	}

	// Build user response
	userResp := &dto.UserResponse{
		ID:         user.ID,
		Email:      email,
		FullName:   user.Name.GetOrZero(),
		Phone:      user.Phone.GetOrZero(),
		IsVerified: user.IsVerified.GetOrZero(),
	}

	return &dto.LoginResponse{
		User:                  userResp,
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewAuthenticationError("Invalid refresh token", exceptions.AuthenticationInvalidToken)
		}
		log.Errorf("failed to find refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
	}

	// Check if already revoked
	if token.RevokedAt.IsValue() {
		return nil, exceptions.NewAuthenticationError("Refresh token has been revoked", exceptions.AuthenticationTokenRevoked)
	}

	// Check if expired
	if token.ExpiresAt.Before(time.Now()) {
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

	// Token rotation: revoke old token and record which token replaced it
	err = service.tokenRepository.ReplaceToken(ctx, req.RefreshToken, newRefreshToken)
	if err != nil {
		log.Errorf("failed to rotate refresh token: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to rotate refresh token"), nil)
	}

	// Create new refresh token — carry over device info from the old token
	deviceName := token.DeviceName.GetOrZero()
	if deviceName == "" {
		deviceName = "Unknown Device"
	}

	tokenSetter := &models.RefreshTokenSetter{
		UserID:     omit.From(token.UserID),
		Token:      omit.From(newRefreshToken),
		DeviceName: omitnull.From(deviceName),
		DeviceID:   omitnull.From(token.DeviceID.GetOrZero()),
		IPAddress:  omitnull.From(ipAddress),
		UserAgent:  omitnull.From(userAgent),
		ExpiresAt:  omit.From(refreshExpiresAt),
	}

	_, err = service.tokenRepository.Create(ctx, tokenSetter)
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
		if errors.Is(err, sql.ErrNoRows) {
			return exceptions.NewAuthenticationError("Invalid refresh token", exceptions.AuthenticationInvalidToken)
		}
		log.Errorf("failed to find refresh token: %s", err)
		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
	}

	// Check if already revoked
	if token.RevokedAt.IsValue() {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
		}
		log.Errorf("failed to find user by ID: %s", err)
		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to get user"), nil)
	}

	// Build user response
	email := ""
	if user.Email.IsValue() {
		email = user.Email.GetOrZero()
	}

	userResp := &dto.UserResponse{
		ID:         user.ID,
		Email:      email,
		FullName:   user.Name.GetOrZero(),
		Phone:      user.Phone.GetOrZero(),
		IsVerified: user.IsVerified.GetOrZero(),
	}

	return &dto.GetMeResponse{
		User:           userResp,
	}, nil
}
