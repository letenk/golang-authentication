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

	accessToken, accessExpiresAt, err := service.jwtConfig.GenerateAccessToken(user.ID, email)
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

// // RefreshToken generates new access token using refresh token
// func (service *AuthServiceImpl) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
// 	// Find refresh token
// 	tokenRecord, err := service.tokenRepository.FindByToken(ctx, req.RefreshToken)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, exceptions.NewAuthenticationError("Invalid or expired refresh token", exceptions.AuthenticationInvalidToken)
// 		}
// 		log.Errorf("failed to find refresh token: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
// 	}

// 	// Check if token is revoked
// 	if tokenRecord.RevokedAt.IsValue() {
// 		return nil, exceptions.NewAuthenticationError("Token has been revoked or already used", exceptions.AuthenticationTokenRevoked)
// 	}

// 	// Check if token is expired
// 	if time.Now().After(tokenRecord.ExpiresAt) {
// 		return nil, exceptions.NewAuthenticationError("Invalid or expired refresh token", exceptions.AuthenticationTokenExpired)
// 	}

// 	// Find user
// 	user, err := service.userRepository.FindByID(ctx, tokenRecord.UserID)
// 	if err != nil {
// 		log.Errorf("failed to find user: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find user"), nil)
// 	}

// 	// Generate new access token
// 	email := ""
// 	if user.Email.IsValue() {
// 		email = user.Email.GetOrZero()
// 	}

// 	accessToken, accessExpiresAt, err := service.jwtConfig.GenerateAccessToken(user.ID, email)
// 	if err != nil {
// 		log.Errorf("failed to generate access token: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to generate access token"), nil)
// 	}

// 	// Generate new refresh token (rotation)
// 	newRefreshToken, newRefreshExpiresAt := service.jwtConfig.GenerateRefreshToken()

// 	// Create new refresh token record using Setter
// 	newTokenSetter := &models.RefreshTokenSetter{
// 		UserID:     omit.From(user.ID),
// 		Token:      omit.From(newRefreshToken),
// 		DeviceName: tokenRecord.DeviceName,
// 		DeviceID:   tokenRecord.DeviceID,
// 		IPAddress:  tokenRecord.IPAddress,
// 		UserAgent:  tokenRecord.UserAgent,
// 		ExpiresAt:  omit.From(newRefreshExpiresAt),
// 	}

// 	_, err = service.tokenRepository.Create(ctx, newTokenSetter)
// 	if err != nil {
// 		log.Errorf("failed to create new refresh token: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataCreateFailed, errors.New("failed to create refresh token"), nil)
// 	}

// 	// Revoke old refresh token
// 	err = service.tokenRepository.Revoke(ctx, req.RefreshToken)
// 	if err != nil {
// 		log.Errorf("failed to revoke old refresh token: %s", err)
// 	}

// 	return &dto.RefreshTokenResponse{
// 		AccessToken:           accessToken,
// 		RefreshToken:          newRefreshToken,
// 		AccessTokenExpiresAt:  accessExpiresAt.Format(time.RFC3339),
// 		RefreshTokenExpiresAt: newRefreshExpiresAt.Format(time.RFC3339),
// 	}, nil
// }

// // Logout revokes a single refresh token
// func (service *AuthServiceImpl) Logout(ctx context.Context, req *dto.LogoutRequest, userID int64) error {
// 	// Find refresh token
// 	tokenRecord, err := service.tokenRepository.FindByToken(ctx, req.RefreshToken)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return exceptions.NewAuthenticationError("Invalid refresh token", exceptions.AuthenticationInvalidToken)
// 		}
// 		log.Errorf("failed to find refresh token: %s", err)
// 		return exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find refresh token"), nil)
// 	}

// 	// Verify token belongs to user
// 	if tokenRecord.UserID != userID {
// 		return exceptions.NewAuthenticationError("Unauthorized", exceptions.AuthenticationUnauthenticated)
// 	}

// 	// Revoke token
// 	err = service.tokenRepository.Revoke(ctx, req.RefreshToken)
// 	if err != nil {
// 		log.Errorf("failed to revoke refresh token: %s", err)
// 		return exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to revoke token"), nil)
// 	}

// 	return nil
// }

// // LogoutAll revokes all refresh tokens for a user
// func (service *AuthServiceImpl) LogoutAll(ctx context.Context, userID int64) (*dto.LogoutAllResponse, error) {
// 	count, err := service.tokenRepository.RevokeAllByUserID(ctx, userID)
// 	if err != nil {
// 		log.Errorf("failed to revoke all refresh tokens: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, errors.New("failed to revoke all tokens"), nil)
// 	}

// 	return &dto.LogoutAllResponse{
// 		Message:             "Logged out from all devices successfully",
// 		RevokedSessionCount: count,
// 	}, nil
// }

// // GetMe returns current user info with active sessions
// func (service *AuthServiceImpl) GetMe(ctx context.Context, userID int64) (*dto.GetMeResponse, error) {
// 	// Find user
// 	user, err := service.userRepository.FindByID(ctx, userID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return nil, exceptions.NewBusinessLogicError(exceptions.DataNotFound, errors.New("user not found"), nil)
// 		}
// 		log.Errorf("failed to find user: %s", err)
// 		return nil, exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find user"), nil)
// 	}

// 	// Build user response
// 	email := ""
// 	if user.Email.IsValue() {
// 		email = user.Email.GetOrZero()
// 	}

// 	userResp := &dto.UserResponse{
// 		ID:         user.ID,
// 		Email:      email,
// 		FullName:   user.Name.GetOrZero(),
// 		Phone:      user.Phone.GetOrZero(),
// 		IsVerified: user.IsVerified.GetOrZero(),
// 		IsActive:   user.IsActive.GetOrZero(),
// 	}

// 	// Get active sessions
// 	activeSessions, err := service.tokenRepository.FindActiveByUserID(ctx, userID)
// 	if err != nil {
// 		log.Errorf("failed to get active sessions: %s", err)
// 		// Don't fail the request, just return empty sessions
// 		activeSessions = []*models.RefreshToken{}
// 	}

// 	// Build session responses
// 	sessionResponses := make([]*dto.ActiveSessionResponse, 0, len(activeSessions))
// 	for _, session := range activeSessions {
// 		lastUsed := session.CreatedAt.GetOrZero()
// 		sessionResp := &dto.ActiveSessionResponse{
// 			ID:         session.ID,
// 			DeviceName: session.DeviceName.GetOrZero(),
// 			DeviceID:   session.DeviceID.GetOrZero(),
// 			IPAddress:  session.IPAddress.GetOrZero(),
// 			UserAgent:  session.UserAgent.GetOrZero(),
// 			LastUsed:   lastUsed.Format(time.RFC3339),
// 			CreatedAt:  lastUsed.Format(time.RFC3339),
// 		}
// 		sessionResponses = append(sessionResponses, sessionResp)
// 	}

// 	return &dto.GetMeResponse{
// 		User:           userResp,
// 		ActiveSessions: sessionResponses,
// 	}, nil
// }
