package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// IsValidPassword validates password with the following requirements:
// - At least one lowercase letter (a-z)
// - At least one uppercase letter (A-Z)
// - At least one digit (0-9)
// - At least one special/unique character (non-alphanumeric)
// Note: Go's regexp doesn't support lookahead, so we check each requirement separately
func IsValidPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check for at least one lowercase letter
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return false
	}

	// Check for at least one digit
	hasDigit, _ := regexp.MatchString(`\d`, password)
	if !hasDigit {
		return false
	}

	// Check for at least one special character (non-alphanumeric)
	hasSpecial, _ := regexp.MatchString(`[^a-zA-Z\d]`, password)
	if !hasSpecial {
		return false
	}

	return true
}