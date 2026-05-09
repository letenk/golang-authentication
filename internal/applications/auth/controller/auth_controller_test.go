package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aarondl/opt/null"
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/bob/models"
	customValidator "github.com/letenk/golang-authentication/configs/validator"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/letenk/golang-authentication/internal/applications/auth/controller"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	authMocks "github.com/letenk/golang-authentication/internal/applications/auth/service/mocks"
	"github.com/letenk/golang-authentication/internal/utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Helpers ---

// newTestEcho creates an Echo instance configured the same as production (validator + error handler).
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = customValidator.NewCustomValidator()
	customValidator.SetupGlobalHttpUnhandleErrors(e)
	return e
}

// stubAuthMiddleware injects a fixed userID into the request context,
// simulating what the authentication middleware does in production.
func stubAuthMiddleware(userID int64) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			h := &headers.Header{UserID: userID}
			reqCtx := context.WithValue(ctx.Request().Context(), headers.ContextHeaders, h)
			ctx.SetRequest(ctx.Request().WithContext(reqCtx))
			return next(ctx)
		}
	}
}

// setupRoutes registers all auth routes on the given echo instance with the provided mock service.
func setupRoutes(e *echo.Echo, mockSvc *authMocks.MockAuthService) {
	c := controller.NewAuthController(mockSvc)
	g := e.Group("/api/v1/auth")
	g.POST("/register", c.Register)
	g.POST("/login", c.Login)
	g.POST("/refresh", c.RefreshToken)
	g.POST("/logout", c.Logout)
	g.POST("/logout-all", c.LogoutAll, stubAuthMiddleware(1))
	g.GET("/me", c.GetMe, stubAuthMiddleware(1))
	g.DELETE("/me", c.DeleteMe, stubAuthMiddleware(1))
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(b)
}

// --- Register ---

func TestAuthController_Register(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success register",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password1!",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Register(mock.Anything, mock.Anything).Return(&models.User{
					ID:    1,
					Email: null.From("test@example.com"),
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "validation error - missing email and phone",
			body: map[string]string{
				"password": "Password1!",
			},
			setupMock:  func(ms *authMocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - invalid email format",
			body: map[string]string{
				"email":    "not-an-email",
				"password": "Password1!",
			},
			setupMock:  func(ms *authMocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "email already registered",
			body: map[string]string{
				"email":    "duplicate@example.com",
				"password": "Password1!",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Register(mock.Anything, mock.Anything).Return(nil, exceptions.NewBusinessLogicError(exceptions.DataConflict, nil, nil))
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- Login ---

func TestAuthController_Login(t *testing.T) {
	successResp := &dto.LoginResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	tests := []struct {
		name       string
		body       any
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success login via email",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "Password1!",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Login(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(successResp, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "validation error - missing email and phone",
			body: map[string]string{
				"password": "Password1!",
			},
			setupMock:  func(ms *authMocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			body: map[string]string{
				"email":    "test@example.com",
				"password": "WrongPassword1!",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Login(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, exceptions.NewAuthenticationError("Invalid credentials", exceptions.AuthenticationInvalidCredentials))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- RefreshToken ---

func TestAuthController_RefreshToken(t *testing.T) {
	successResp := &dto.RefreshTokenResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}

	tests := []struct {
		name       string
		body       any
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success refresh token",
			body: map[string]string{
				"refresh_token": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().RefreshToken(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(successResp, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "validation error - missing refresh token",
			body:       map[string]string{},
			setupMock:  func(ms *authMocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "token revoked",
			body: map[string]string{
				"refresh_token": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().RefreshToken(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, exceptions.NewAuthenticationError("Refresh token has been revoked", exceptions.AuthenticationTokenRevoked))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "token expired",
			body: map[string]string{
				"refresh_token": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().RefreshToken(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, exceptions.NewAuthenticationError("Refresh token has expired", exceptions.AuthenticationTokenExpired))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", jsonBody(t, tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- Logout ---

func TestAuthController_Logout(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success logout",
			body: map[string]string{
				"refresh_token": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Logout(mock.Anything, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "validation error - missing refresh token",
			body:       map[string]string{},
			setupMock:  func(ms *authMocks.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "token already revoked",
			body: map[string]string{
				"refresh_token": "550e8400-e29b-41d4-a716-446655440000",
			},
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().Logout(mock.Anything, mock.Anything).Return(exceptions.NewAuthenticationError("Refresh token has already been revoked", exceptions.AuthenticationTokenRevoked))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", jsonBody(t, tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- LogoutAll ---

func TestAuthController_LogoutAll(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success logout all devices",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().LogoutAll(mock.Anything, int64(1)).Return(3, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().LogoutAll(mock.Anything, int64(1)).Return(0, exceptions.NewBusinessLogicError(exceptions.DataUpdateFailed, nil, nil))
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- DeleteMe ---

func TestAuthController_DeleteMe(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success delete account",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().DeleteAccount(mock.Anything, int64(1)).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().DeleteAccount(mock.Anything, int64(1)).Return(exceptions.NewBusinessLogicError(exceptions.DataDeleteFailed, nil, nil))
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/me", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

// --- GetMe ---

func TestAuthController_GetMe(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(ms *authMocks.MockAuthService)
		wantStatus int
	}{
		{
			name: "success get current user",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().GetMe(mock.Anything, int64(1)).Return(&dto.GetMeResponse{
					User: &dto.UserResponse{ID: 1, Email: "test@example.com"},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "user not found",
			setupMock: func(ms *authMocks.MockAuthService) {
				ms.EXPECT().GetMe(mock.Anything, int64(1)).Return(nil, exceptions.NewBusinessLogicError(exceptions.DataNotFound, nil, nil))
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			mockSvc := authMocks.NewMockAuthService(t)
			tt.setupMock(mockSvc)
			setupRoutes(e, mockSvc)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
