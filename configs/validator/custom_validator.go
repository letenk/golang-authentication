package validator

import (
	"github.com/go-playground/validator/v10"
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
		// Return the original validation error without wrapping it
		// This allows GlobalUnHandleErrors to format it properly as a map
		return err
	}
	return nil
}
