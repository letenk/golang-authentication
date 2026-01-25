package dto

// Validate custom validation for RegisterRequest
func (r RegisterRequest) Validate() error {
	// At least email or phone must be provided
	if r.Email == "" && r.Phone == "" {
		return NewValidationError("email or phone", "at least one of email or phone is required")
	}
	return nil
}

// ValidationError untuk custom error message
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}
