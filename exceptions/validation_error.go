package exceptions

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func NewValidationError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}
