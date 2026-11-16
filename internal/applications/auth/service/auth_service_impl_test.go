package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	"github.com/letenk/golang-authentication/internal/applications/auth/service"
	emailSvcMocks "github.com/letenk/golang-authentication/internal/applications/email/service/mocks"
	otpRepoMocks "github.com/letenk/golang-authentication/internal/applications/otp/repository/db/mocks"
	refreshMocks "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db/mocks"
	trxMocks "github.com/letenk/golang-authentication/internal/applications/transaction/mocks"
	userMocks "github.com/letenk/golang-authentication/internal/applications/user/repository/db/mocks"
	"github.com/letenk/golang-authentication/internal/domain"
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
	mockOTP := otpRepoMocks.NewMockOTPRepository(t)
	mockEmail := emailSvcMocks.NewMockEmailService(t)
	mockTrx := trxMocks.NewMockTrxService(t)
	otpConfig := &credential.OTPConfig{Expire: "5m", Length: 6}

	// Allow goroutine side-effect calls from Register's async SendVerificationEmail
	// for methods that are not explicitly tested. FindByID is excluded here because
	// other tests (e.g. GetMe) register their own specific expectations for it.
	mockTrx.On("WithTx", mock.Anything, mock.Anything).Maybe().Return(nil)
	mockEmail.On("SendVerificationOTP", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

	svc := service.NewAuthService(mockTrx, mockUser, mockToken, mockOTP, mockEmail, newTestJWTConfig(t), otpConfig)
	return svc, mockUser, mockToken
}

func newTestUser(id int64, email, phone, hashedPassword string) *domain.User {
	u := &domain.User{
		ID:        id,
		Email:     email,
		Phone:     phone,
		Password:  hashedPassword,
		LoginType: "email",
	}
	if phone != "" {
		if email == "" {
			u.LoginType = "phone"
		}
	}
	return u
}

func newActiveToken(userID int64, token string) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:         1,
		UserID:     userID,
		Token:      token,
		DeviceName: "Test Device",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}
}

func newRevokedToken(userID int64, token string) *domain.RefreshToken {
	rt := newActiveToken(userID, token)
	revokedAt := time.Now()
	rt.RevokedAt = &revokedAt
	return rt
}

func newExpiredToken(userID int64, token string) *domain.RefreshToken {
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
		wantCode  int
	}{
		{
			name: "success register via email",
			param: &dto.ParameterRegister{
				FullName: "Test User",
				Email:    "test@example.com",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "test@example.com").Return(nil, domain.ErrNotFound)
				mu.EXPECT().Create(mock.Anything, mock.Anything).Return(newTestUser(1, "test@example.com", "", "hashed"), nil)
				// Allow async goroutine (SendVerificationEmail) to call FindByID.
				// Returns ErrNoRows so the goroutine exits early without further calls.
				mu.On("FindByID", mock.Anything, mock.Anything).Maybe().Return(nil, domain.ErrNotFound)
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
				mu.EXPECT().FindByPhone(mock.Anything, "+628123456789").Return(nil, domain.ErrNotFound)
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
				mu.EXPECT().FindByPhone(mock.Anything, "+628123456789").Return(newTestUser(1, "", "+628123456789", "hashed"), nil)
			},
			wantErr: true,
			errType: &exceptions.BusinessLogicError{},
		},
		{
			// Pre-check passes (concurrent signup, or the email belongs to a
			// soft-deleted account) but the unique index rejects the insert.
			name: "email taken at insert",
			param: &dto.ParameterRegister{
				Email:    "race@example.com",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByEmail(mock.Anything, "race@example.com").Return(nil, domain.ErrNotFound)
				mu.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, domain.ErrEmailTaken)
			},
			wantErr:  true,
			wantCode: exceptions.DataConflict,
		},
		{
			name: "phone taken at insert",
			param: &dto.ParameterRegister{
				Phone:    "+628123456789",
				Password: "Password1!",
			},
			setupMock: func(mu *userMocks.MockUserRepository) {
				mu.EXPECT().FindByPhone(mock.Anything, "+628123456789").Return(nil, domain.ErrNotFound)
				mu.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, domain.ErrPhoneTaken)
			},
			wantErr:  true,
			wantCode: exceptions.DataConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockUser, _ := newTestService(t)
			tt.setupMock(mockUser)

			result, err := svc.Register(context.Background(), tt.param)

			// Allow async goroutine (SendVerificationEmail) to complete before mock cleanup.
			time.Sleep(20 * time.Millisecond)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.As(err, &tt.errType))
				}
				if tt.wantCode != 0 {
					var bizErr *exceptions.BusinessLogicError
					if assert.True(t, errors.As(err, &bizErr)) {
						assert.Equal(t, tt.wantCode, bizErr.ErrorCode)
					}
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
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&domain.RefreshToken{}, nil)
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
				mu.EXPECT().FindByPhone(mock.Anything, "+628123456789").Return(newTestUser(1, "", "+628123456789", hashedPassword), nil)
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&domain.RefreshToken{}, nil)
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
				mu.EXPECT().FindByEmail(mock.Anything, "notfound@example.com").Return(nil, domain.ErrNotFound)
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
				mt.EXPECT().Create(mock.Anything, mock.Anything).Return(&domain.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "token not found",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(nil, domain.ErrNotFound)
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
		{
			// Another request rotated the same token between FindByToken and
			// ReplaceToken. No new token may be issued (the strict mock fails
			// on an unexpected Create call).
			name: "token rotated concurrently",
			setupMock: func(mt *refreshMocks.MockRefreshTokenRepository) {
				mt.EXPECT().FindByToken(mock.Anything, oldToken).Return(newActiveToken(1, oldToken), nil)
				mt.EXPECT().ReplaceToken(mock.Anything, oldToken, mock.Anything).Return(domain.ErrNotFound)
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
				mt.EXPECT().FindByToken(mock.Anything, token).Return(nil, domain.ErrNotFound)
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
				mu.EXPECT().FindByID(mock.Anything, int64(99)).Return(nil, domain.ErrNotFound)
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
