package validator

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/exceptions"
)

// MapperErrorCode maps various error types to HTTP codes and business error codes
// Returns validation errors as map[string][]string if it's a validation error
func MapperErrorCode(err error) (errHttpCode int, errBusinessCode int, errMessage string, errUnexpected error, validationErrors map[string][]string) {

	// Handle echo.HTTPError (should be handled in GlobalUnHandleErrors, but this is a safety net)
	if echoErr, ok := err.(*echo.HTTPError); ok {
		statusCode := echoErr.Code
		message := ""

		if msg, ok := echoErr.Message.(string); ok {
			message = msg
		} else {
			message = http.StatusText(statusCode)
		}

		return statusCode, statusCode, message, nil, nil
	}

	// Handle BusinessLogicError
	if errors.Is(err, exceptions.TargetBusinessLogicError) {
		businessLogicError := err.(*exceptions.BusinessLogicError)
		errorCode := businessLogicError.ErrorCode
		errorLogic := exceptions.BusinessLogicReason(errorCode)

		// Format business logic error as map for consistency
		var errorMsg string
		if businessLogicError.Err != nil {
			errorMsg = businessLogicError.Err.Error()
		} else {
			errorMsg = errorLogic.Message
		}

		// Create errors map with general key or specific field if Data contains field name
		errorMap := map[string][]string{
			"message": {errorMsg},
		}

		return errorLogic.HttpCode, errorLogic.ErrCode, errorMsg, nil, errorMap
	}

	// Handle AuthenticationError
	if authErr, ok := err.(*exceptions.AuthenticationError); ok {
		errorLogic := exceptions.AuthenticationReason(authErr.Code)

		// Use custom message if provided, otherwise use default from mapping
		errorMsg := authErr.Message
		if errorMsg == "" {
			errorMsg = errorLogic.Message
		}

		// Create errors map for consistency
		errorMap := map[string][]string{
			"message": {errorMsg},
		}

		return errorLogic.HttpCode, errorLogic.ErrCode, errorMsg, nil, errorMap
	}

	// Handle custom field ValidationError (e.g. "email or phone required")
	if valErr, ok := err.(exceptions.ValidationError); ok {
		errorMap := map[string][]string{
			valErr.Field: {valErr.Message},
		}
		return http.StatusBadRequest, http.StatusBadRequest, "Validation Failed", nil, errorMap
	}

	// Handle ValidationErrors from validator
	if _, ok := err.(validator.ValidationErrors); ok {
		// Return validation errors as structured map
		validationErrorsMap := FormatValidationErrorMap(err)
		return http.StatusBadRequest, http.StatusBadRequest, "Validation Failed", nil, validationErrorsMap
	}

	// Default to internal server error
	errorMessage := err.Error()
	return http.StatusInternalServerError, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), errors.New(errorMessage), nil
}
