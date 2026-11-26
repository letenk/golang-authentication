package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func preflight(t *testing.T, allowed []string, origin string) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	e.Use(CORS(allowed))
	e.GET("/api/v1/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/health", nil)
	req.Header.Set(echo.HeaderOrigin, origin)
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// A listed origin may call the API with cookies (credentials).
func TestCORS_AllowsListedOriginWithCredentials(t *testing.T) {
	rec := preflight(t, []string{"http://localhost:3000"}, "http://localhost:3000")

	assert.Equal(t, "http://localhost:3000", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", rec.Header().Get(echo.HeaderAccessControlAllowCredentials))
}

func TestCORS_RejectsUnlistedOrigin(t *testing.T) {
	rec := preflight(t, []string{"http://localhost:3000"}, "https://evil.example.com")

	assert.Empty(t, rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
}
