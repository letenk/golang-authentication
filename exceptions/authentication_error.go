package exceptions

type AuthenticationError struct {
	Message string
	Code    AuthenticationCode
}

func (error *AuthenticationError) Error() string {
	return error.Message
}

func NewAuthenticationError(message string, errorCode AuthenticationCode) error {
	return &AuthenticationError{
		Message: message,
		Code:    errorCode,
	}
}
