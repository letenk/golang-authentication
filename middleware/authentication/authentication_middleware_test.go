package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	customValidator "github.com/letenk/golang-authentication/configs/validator"
	authnService "github.com/letenk/golang-authentication/internal/applications/authentication/service"
	middleware "github.com/letenk/golang-authentication/middleware/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = customValidator.NewCustomValidator()
	customValidator.SetupGlobalHttpUnhandleErrors(e)
	return e
}

// newJWTConfig builds a JWTConfig; accessExpire accepts negative durations (e.g. "-1m")
// to mint an already-expired access token.
func newJWTConfig(t *testing.T, accessExpire string) *jwt_config.JWTConfig {
	t.Helper()
	cfg, err := jwt_config.NewJWTConfig(&credential.JWTConfig{
		Secret:             "test-secret",
		AccessTokenExpire:  accessExpire,
		RefreshTokenExpire: "7d",
	})
	require.NoError(t, err)
	return cfg
}

func TestAuthenticationMiddleware_Authenticate(t *testing.T) {
	tests := []struct {
		name       string
		setupReq   func(t *testing.T, req *http.Request) // attach token to header or cookie
		wantStatus int
	}{
		{
			name: "valid token via cookie passes through (200)",
			setupReq: func(t *testing.T, req *http.Request) {
				cfg := newJWTConfig(t, "15m")
				token, _, err := cfg.GenerateAccessToken(1)
				require.NoError(t, err)
				req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid token via Authorization header passes through (200)",
			setupReq: func(t *testing.T, req *http.Request) {
				cfg := newJWTConfig(t, "15m")
				token, _, err := cfg.GenerateAccessToken(1)
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantStatus: http.StatusOK,
		},
		{
			// Regression: an expired access token must be 401 (not 400) so the client
			// can transparently refresh and replay the request.
			name: "expired token is unauthorized (401)",
			setupReq: func(t *testing.T, req *http.Request) {
				cfg := newJWTConfig(t, "-1m") // expires in the past → already expired
				token, _, err := cfg.GenerateAccessToken(1)
				require.NoError(t, err)
				req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed token is unauthorized (401)",
			setupReq: func(t *testing.T, req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "access_token", Value: "not-a-jwt"})
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing token (no header, no cookie) is unauthorized (401)",
			setupReq:   func(t *testing.T, req *http.Request) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatorCfg := newJWTConfig(t, "15m")
			authSvc := authnService.NewAuthenticationService(validatorCfg)
			authMw := middleware.NewAuthenticationMiddleware(authSvc)

			e := newTestEcho()
			e.GET("/protected", func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}, authMw.Authenticate(true))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			tt.setupReq(t, req)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
