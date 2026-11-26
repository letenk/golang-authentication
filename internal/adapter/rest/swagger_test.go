package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func get(e *echo.Echo, path string) int {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code
}

func TestRegisterSwagger_ServesDocsOutsideProduction(t *testing.T) {
	for _, env := range []string{"development", "staging"} {
		t.Run(env, func(t *testing.T) {
			e := echo.New()
			RegisterSwagger(e, env)

			assert.Equal(t, http.StatusOK, get(e, "/swagger/index.html"))
			assert.Equal(t, http.StatusOK, get(e, "/swagger/doc.json"))
		})
	}
}

// In production the route must not exist at all, not just be hidden.
func TestRegisterSwagger_DoesNotServeDocsInProduction(t *testing.T) {
	e := echo.New()
	RegisterSwagger(e, "production")

	assert.Equal(t, http.StatusNotFound, get(e, "/swagger/index.html"))
	assert.Equal(t, http.StatusNotFound, get(e, "/swagger/doc.json"))
}
