package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBase_TableDriven(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		httpCode       int
		errorCode      int
		message        string
		data           interface{}
		err            error
		expectHTTPCode int
		expectError    string
		expectDataNil  bool
	}{
		{
			name:           "success with data",
			httpCode:       http.StatusOK,
			errorCode:      http.StatusOK,
			message:        "OK",
			data:           map[string]string{"foo": "bar"},
			err:            nil,
			expectHTTPCode: http.StatusOK,
			expectError:    "",
			expectDataNil:  false,
		},
		{
			name:           "error response",
			httpCode:       http.StatusBadRequest,
			errorCode:      http.StatusBadRequest,
			message:        "Bad Request",
			data:           nil,
			err:            errors.New("invalid input"),
			expectHTTPCode: http.StatusBadRequest,
			expectError:    "invalid input",
			expectDataNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := Base(ctx, tt.httpCode, tt.errorCode, tt.message, tt.data, tt.err)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectHTTPCode, rec.Code)
			assert.NotEmpty(t, rec.Header().Get("date"))

			var resp map[string]interface{}
			json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&resp)

			assert.Equal(t, float64(tt.errorCode), resp["code"])
			assert.Equal(t, tt.message, resp["message"])
			assert.NotEmpty(t, resp["serverTime"])

			if tt.expectDataNil {
				assert.Nil(t, resp["data"])
			} else {
				assert.NotNil(t, resp["data"])
			}

			if tt.expectError != "" {
				assert.Equal(t, tt.expectError, resp["error"])
			}
		})
	}
}

func TestSuccess_PanicWhenDataNil(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	assert.Panics(t, func() {
		_ = Success(ctx, nil)
	})
}

func TestSuccessWithMessage_PanicWhenDataNil(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	assert.Panics(t, func() {
		_ = SuccessWithMessage(ctx, "custom message", nil)
	})
}

func TestError_TableDriven(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		httpCode   int
		err        error
		expectCode int
	}{
		{
			name:       "not found error",
			httpCode:   http.StatusNotFound,
			err:        errors.New("resource not found"),
			expectCode: http.StatusNotFound,
		},
		{
			name:       "internal server error",
			httpCode:   http.StatusInternalServerError,
			err:        errors.New("internal error"),
			expectCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			err := Error(ctx, tt.httpCode, tt.err)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectCode, rec.Code)

			var resp map[string]interface{}
			json.NewDecoder(rec.Body).Decode(&resp)

			assert.Equal(t, float64(tt.expectCode), resp["code"])
			assert.Equal(t, tt.err.Error(), resp["error"])
		})
	}
}
