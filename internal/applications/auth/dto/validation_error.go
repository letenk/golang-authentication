package dto

import "github.com/letenk/golang-authentication/exceptions"

// Validate custom validation for RegisterRequest
func (r RegisterRequest) Validate() error {
	// At least email or phone must be provided
	if r.Email == "" && r.Phone == "" {
		return exceptions.NewValidationError("email or phone", "at least one of email or phone is required")
	}
	return nil
}

// Validate custom validation for LoginRequest
func (r LoginRequest) Validate() error {
	// At least email or phone must be provided
	if r.Email == "" && r.Phone == "" {
		return exceptions.NewValidationError("email or phone", "at least one of email or phone is required")
	}
	return nil
}
