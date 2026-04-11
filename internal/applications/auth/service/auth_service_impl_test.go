package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/aarondl/opt/null"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
	emailSvcMocks "github.com/letenk/golang-authentication/internal/applications/email/service/mocks"
	otpRepoMocks "github.com/letenk/golang-authentication/internal/applications/password_reset/repository/db/mocks"
	refreshMocks "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db/mocks"
	trxMocks "github.com/letenk/golang-authentication/internal/applications/transaction/mocks"
	userMocks "github.com/letenk/golang-authentication/internal/applications/user/repository/db/mocks"
	"github.com/letenk/golang-authentication/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Helpers ---

func newTestJWTConfig(t *testing.T) *jwt_config.JWTConfig {
	t.Helper()
	cfg := &credential.JWTConfig{
		Secret:             "test-secret-key",
		AccessTokenExpire:  "15m",
		RefreshTokenExpire: "7d",
	}
	jwtCfg, err := jwt_config.NewJWTConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create test JWT config: %v", err)
	}
	return jwtCfg
}

func newTestService(t *testing.T) (*service.AuthServiceImpl, *userMocks.MockUserRepository, *refreshMocks.MockRefreshTokenRepository) {
	t.Helper()
	mockUser := userMocks.NewMockUserRepository(t)
	mockToken := refreshMocks.NewMockRefreshTokenRepository(t)
	mockOTP := otpRepoMocks.NewMockPasswordResetOTPRepository(t)
	mockEmail := emailSvcMocks.NewMockEmailService(t)
	mockTrx := trxMocks.NewMockTrxService(t)
	otpConfig := &credential.OTPConfig{Expire: "5m", Length: 6}
	svc := service.NewAuthService(mockTrx, mockUser, mockToken, mockOTP, mockEmail, newTestJWTConfig(t), otpConfig)
	return svc, mockUser, mockToken
}

func newTestUser(id int64, email, phone, hashedPassword string) *models.User {
	u := &models.User{
		ID:         id,
		Password:   hashedPassword,
		LoginType:  "email",
		IsVerified: null.From(false),
	}
	if email != "" {
		u.Email = null.From(email)
	}
	if phone != "" {
		u.Phone = null.From(phone)
		if email == "" {
			u.LoginType = "phone"
		}
	}
	return u
}

func newActiveToken(userID int64, token string) *models.RefreshToken {
	return &models.RefreshToken{
		ID:         1,
		UserID:     userID,
		Token:      token,
		DeviceName: null.From("Test Device"),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
}

func newRevokedToken(userID int64, token string) *models.RefreshToken {
	rt := newActiveToken(userID, token)
	rt.RevokedAt = null.From(time.Now())
	return rt
}

func newExpiredToken(userID int64, token string) *models.RefreshToken {
	rt := newActiveToken(userID, token)
	rt.ExpiresAt = time.Now().Add(-1 * time.Hour)
	return rt
}

// --- Register ---

func TestAuthService_Register(t *testing.T) {
	tests := []struct {
		name      string
		param     *dto.ParameterRegister
		setupMock func(mu *userMocks.MockUserRepository)
		wantErr   bool
		errType   interface{}
	}{
		{
			name: "success register via email",
			param: &dto.ParameterRegister{
				FullName: "Test User",
				Email:    "test@example.com",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
				mu.EXPECT().Create(mock.Anything, mock.Anything).Return(newTestUser(1, "test@example.com", "", "hashed"), nil)
			},
			wantErr: false,
		},
		{
			name: "success register via phone",
			param: &dto.ParameterRegister{
				FullName: "Test User",
				Phone:    "+628123456789",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByPhone(mock.Anything, "628123456789").Return(nil, sql.ErrNoRows)
				mu.EXPECT().Create(mock.Anything, mock.Anything).Return(newTestUser(1, "", "+628123456789", "hashed"), nil)
			},
			wantErr: false,
		},
		{
			name: "email already registered",
			param: &dto.ParameterRegister{
				Email:    "duplicate@example.com",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "duplicate@example.com").Return(newTestUser(1, "duplicate@example.com", "", "hashed"), nil)
			},
			wantErr: true,
			errType: &exceptions.BusinessLogicError{},
		},
		{
			name: "phone already registered",
			param: &dto.ParameterRegister{
				Phone:    "+628123456789",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByPhone(mock.Anything, "628123456789").Return(newTestUser(1, "", "628123456789", "hashed"), nil)
			},
			wantErr: true,
			errType: &exceptions.BusinessLogicError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockUser, _ := newTestService(t)
			tt.setupMock(mockUser)

			result, err := svc.Register(context.Background(), tt.param)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.As(err, &tt.errType))
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// --- Login ---

func TestAuthService_Login(t *testing.T) {
	const testPassword = "Password1!"
	hashedPassword, _ := utils.HashPassword(testPassword)

	tests := []struct {
		name      string
		req       *dto.LoginRequest
		setupMock func(mu *userMocks.MockUserRepository, mt *refreshMocks.MockRefreshTokenRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "success login via email",
			req: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: testPassword,
			},
			setupMock: func(mu *userMocks.MockUserRepository, mt *refreshMocks.MockRefreshTokenRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "test@example.com").Return(newTestUser(1, "test@example.com", "", hashedPassword), nil)
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&models.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "success login via phone",
			req: &dto.LoginRequest{
				Phone:    "+628123456789",
				Password: testPassword,
			},
			setupMock: func(mu *userMocks.MockUserRepository, mt *refreshMocks.MockRefreshTokenRepository) {
				mu.EXPECT().FindByPhone(mock.Anything, "628123456789").Return(newTestUser(1, "", "628123456789", hashedPassword), nil)
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&models.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			req: &dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: testPassword,
			},
			setupMock: func(mu *userMocks.MockUserRepository, mt *refreshMocks.MockRefreshTokenRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "notfound@example.com").Return(nil, sql.ErrNoRows)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationInvalidCredentials, authErr.Code)
			},
		},
		{
			name: "wrong password",
			req: &dto.LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository, mt *refreshMocks.MockRefreshTokenRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "test@example.com").Return(newTestUser(1, "test@example.com", "", hashedPassword), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationInvalidCredentials, authErr.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockUser, mockToken := newTestService(t)
			tt.setupMock(mockUser, mockToken)

			result, err := svc.Login(context.Background(), tt.req, "127.0.0.1", "test-agent")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
			}
		})
	}
}

// --- RefreshToken ---

func TestAuthService_RefreshToken(t *testing.T) {
	const oldToken = "old-refresh-token-uuid"

	tests := []struct {
		name      string
		setupMock func(mt *refreshMocks.MockRefreshTokenRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "success token rotation",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(newActiveToken(1, oldToken), nil)
				mt.EXPECT().ReplaceToken(mock.Anything, oldToken, mock.Anything).Return(nil)
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&models.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "token not found",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(nil, sql.ErrNoRows)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationInvalidToken, authErr.Code)
			},
		},
		{
			name: "token already revoked",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(newRevokedToken(1, oldToken), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationTokenRevoked, authErr.Code)
			},
		},
		{
			name: "token expired",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(newExpiredToken(1, oldToken), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationTokenExpired, authErr.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, mockToken := newTestService(t)
			tt.setupMock(mockToken)

			result, err := svc.RefreshToken(context.Background(), &dto.RefreshTokenRequest{RefreshToken: oldToken}, "127.0.0.1", "test-agent")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.AccessToken)
				assert.NotEmpty(t, result.RefreshToken)
				assert.NotEqual(t, oldToken, result.RefreshToken)
			}
		})
	}
}

// --- Logout ---

func TestAuthService_Logout(t *testing.T) {
	const token = "some-refresh-token"

	tests := []struct {
		name      string
		setupMock func(mt *refreshMocks.MockRefreshTokenRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "success logout",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, token).Return(newActiveToken(1, token), nil)
				mt.EXPECT().Revoke(mock.Anything, token).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "token not found",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, token).Return(nil, sql.ErrNoRows)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationInvalidToken, authErr.Code)
			},
		},
		{
			name: "token already revoked",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, token).Return(newRevokedToken(1, token), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var authErr *exceptions.AuthenticationError
				assert.True(t, errors.As(err, &authErr))
				assert.Equal(t, exceptions.AuthenticationTokenRevoked, authErr.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, mockToken := newTestService(t)
			tt.setupMock(mockToken)

			err := svc.Logout(context.Background(), token)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- LogoutAll ---

func TestAuthService_LogoutAll(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		setupMock func(mt *refreshMocks.MockRefreshTokenRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name:   "success revoke all sessions",
			userID: 1,
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().RevokeAllByUserID(mock.Anything, int64(1)).Return(3, nil)
			},
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:   "db error",
			userID: 1,
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().RevokeAllByUserID(mock.Anything, int64(1)).Return(0, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, mockToken := newTestService(t)
			tt.setupMock(mockToken)

			count, err := svc.LogoutAll(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}

// --- GetMe ---

func TestAuthService_GetMe(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		setupMock func(mu *userMocks.MockUserRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:   "success get current user",
			userID: 1,
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByID(mock.Anything, int64(1)).Return(newTestUser(1, "test@example.com", "", "hashed"), nil)
			},
			wantErr: false,
		},
		{
			name:   "user not found",
			userID: 99,
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByID(mock.Anything, int64(99)).Return(nil, sql.ErrNoRows)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				var bizErr *exceptions.BusinessLogicError
				assert.True(t, errors.As(err, &bizErr))
				assert.Equal(t, exceptions.DataNotFound, bizErr.ErrorCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockUser, _ := newTestService(t)
			tt.setupMock(mockUser)

			result, err := svc.GetMe(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.checkErr != nil {
					tt.checkErr(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.User.ID)
			}
		})
	}
}
