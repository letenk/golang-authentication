package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/configs/credential"
	customValidator "github.com/letenk/golang-authentication/configs/validator"
	"github.com/letenk/golang-authentication/internal/applications/auth/dto"
	authMocks "github.com/letenk/golang-authentication/internal/applications/auth/service/mocks"
	"github.com/letenk/golang-authentication/internal/applications/user/controller"
	"github.com/letenk/golang-authentication/internal/utils/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	credential.Config = credential.NewAppConfig()
	os.Exit(m.Run())
}

// putProfile sends PUT /api/v1/user/profile as user 1 through an Echo instance
// configured like production (validator + global error handler).
func putProfile(t *testing.T, mockSvc *authMocks.MockAuthService, body string) *httptest.ResponseRecorder {
	t.Helper()

	e := echo.New()
	e.Validator = customValidator.NewCustomValidator()
	customValidator.SetupGlobalHttpUnhandleErrors(e)

	asUser1 := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			reqCtx := context.WithValue(ctx.Request().Context(), headers.ContextHeaders, &headers.Header{UserID: 1})
			ctx.SetRequest(ctx.Request().WithContext(reqCtx))
			return next(ctx)
		}
	}
	e.PUT("/api/v1/user/profile", controller.NewUserController(mockSvc).UpdateProfile, asUser1)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/user/profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// A body with neither full_name nor phone must be rejected before the service
// is called. The mock has no expectations, so any service call fails the test.
func TestUserController_UpdateProfile_RejectsEmptyBody(t *testing.T) {
	rec := putProfile(t, authMocks.NewMockAuthService(t), `{}`)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body struct {
		Errors map[string][]string `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body.Errors, "full_name or phone")
}

func TestUserController_UpdateProfile_AcceptsOneField(t *testing.T) {
	mockSvc := authMocks.NewMockAuthService(t)
	mockSvc.EXPECT().
		UpdateProfile(mock.Anything, int64(1), &dto.UpdateProfileRequest{FullName: "John Doe"}).
		Return(&dto.UserResponse{ID: 1, FullName: "John Doe"}, nil)

	rec := putProfile(t, mockSvc, `{"full_name":"John Doe"}`)

	assert.Equal(t, http.StatusOK, rec.Code)
}
