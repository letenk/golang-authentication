package helper

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type TestRequest struct {
	Name  string `json:"name" validate:"required" mod:"trim"`
	Email string `json:"email" validate:"required,email"`
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}


func TestBindAndValidate(t *testing.T) {
	e := echo.New()
	e.Validator = &CustomValidator{
		validator: validator.New(),
	}

	tests := []struct {
		name       string
		body       string
		wantErr    bool
		assertFunc func(t *testing.T, req *TestRequest)
	}{
		{
			name:    "success valid request",
			body:    `{"name":"  Rjhon  ","email":"rjhon@mail.com"}`,
			wantErr: false,
			assertFunc: func(t *testing.T, req *TestRequest) {
				// memastikan mold (trim) bekerja
				assert.Equal(t, "Rjhon", req.Name)
				assert.Equal(t, "rjhon@mail.com", req.Email)
			},
		},
		{
			name:    "failed validation missing name",
			body:    `{"email":"rjhon@mail.com"}`,
			wantErr: true,
		},
		{
			name:    "failed validation invalid email",
			body:    `{"name":"Rjhon","email":"not-email"}`,
			wantErr: true,
		},
		{
			name:    "failed bind invalid json",
			body:    `{"name":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				bytes.NewBufferString(tt.body),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			var payload TestRequest
			err := BindAndValidate(ctx, &payload)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.assertFunc != nil {
				tt.assertFunc(t, &payload)
			}
		})
	}
}
