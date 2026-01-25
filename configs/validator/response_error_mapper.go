package validator

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/letenk/golang-authentication/exceptions"
)

// MapperErrorCode maps various error types to HTTP codes and business error codes
func MapperErrorCode(err error) (errHttpCode int, errBusinessCode int, errMessage string, errUnexpected error, data any) {

	// Handle BusinessLogicError
	if errors.Is(err, exceptions.TargetBusinessLogicError) {
		businessLogicError := err.(*exceptions.BusinessLogicError)
		errorCode := businessLogicError.ErrorCode
		errorLogic := exceptions.BusinessLogicReason(errorCode)

		return errorLogic.HttpCode, errorLogic.ErrCode, errorLogic.Message, businessLogicError.Err, businessLogicError.Data
	}

	// Handle ValidationErrors from validator
	errorMessage := err.Error()
	if castedObject, ok := err.(validator.ValidationErrors); ok {
		for _, err := range castedObject {
			switch err.Tag() {
			case "required":
				errorMessage = fmt.Sprintf("%s is required", err.Field())
			case "email":
				errorMessage = fmt.Sprintf("%s is not valid email", err.Field())
			case "gte":
				errorMessage = fmt.Sprintf("%s value must be greater than %s", err.Field(), err.Param())
			case "lte":
				errorMessage = fmt.Sprintf("%s value must be lower than %s", err.Field(), err.Param())
			case "password":
				errorMessage = fmt.Sprintf("%s %s", err.Field(), err.Value())
			case "e164":
				errorMessage = fmt.Sprintf("%s must be in valid E.164 format", err.Field())
			}
			break
		}

		return http.StatusBadRequest, http.StatusBadRequest, errorMessage, nil, nil
	}

	// Default to internal server error
	return http.StatusInternalServerError, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), errors.New(errorMessage), nil
}
