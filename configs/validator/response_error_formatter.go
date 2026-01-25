package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError formats validation errors into readable messages
func FormatValidationError(err error) []string {
	var messages []string

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			message := formatFieldError(fe)
			messages = append(messages, message)
		}
	} else {
		messages = append(messages, err.Error())
	}

	return messages
}

func formatFieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
	case "e164":
		return fmt.Sprintf("%s must be in E.164 format (e.g., +628123456789)", field)
	case "strong_password":
		return fmt.Sprintf("%s must contain at least one uppercase letter, one lowercase letter, one digit, and one special character", field)
	default:
		return fmt.Sprintf("%s failed validation for '%s'", field, fe.Tag())
	}
}
