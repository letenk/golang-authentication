package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	
	// Register custom validators
	v.RegisterValidation("strong_password", IsValidPassword)
	
	return &CustomValidator{validator: v}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Format validation errors
		errorMessages := FormatValidationError(err)
		return echo.NewHTTPError(400, strings.Join(errorMessages, "; "))
	}
	return nil
}
