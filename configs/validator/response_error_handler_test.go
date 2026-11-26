package validator

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/exceptions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handle runs err through the global error handler and returns the response.
func handle(t *testing.T, err error) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	rec := httptest.NewRecorder()
	ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	GlobalUnHandleErrors()(err, ctx)
	return rec
}

// An unexpected error must not leak its text (for example a raw database
// error) to the client.
func TestGlobalUnHandleErrors_HidesUnexpectedErrorText(t *testing.T) {
	rec := handle(t, errors.New(`pq: relation "users" does not exist`))

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "pq:")
	assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error"`)
}

// Business errors that map to 500 must not expose their message either.
func TestGlobalUnHandleErrors_HidesServerSideBusinessErrorText(t *testing.T) {
	err := exceptions.NewBusinessLogicError(exceptions.DataGetFailed, errors.New("failed to find user: connection refused"), nil)

	rec := handle(t, err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "connection refused")
	assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error"`)
}

// Client errors keep their message so the caller knows what to fix.
func TestGlobalUnHandleErrors_KeepsClientErrorMessage(t *testing.T) {
	err := exceptions.NewBusinessLogicError(exceptions.DataConflict, errors.New("email already registered"), nil)

	rec := handle(t, err)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "email already registered")
}
