package validator

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/internal/helper/response"
)

func SetupGlobalHttpUnhandleErrors(e *echo.Echo) {
	e.HTTPErrorHandler = GlobalUnHandleErrors()
	log.Infof("initialized GlobalUnHandleErrors : success")
}

func GlobalUnHandleErrors() func(err error, ctx echo.Context) {
	return func(err error, ctx echo.Context) {
		// Handle echo.HTTPError (from Echo framework validation/binding)
		if httpErr, ok := err.(*echo.HTTPError); ok {
			// Extract status code and message from HTTPError
			statusCode := httpErr.Code
			message := ""

			// HTTPError.Message can be string or any type
			if msg, ok := httpErr.Message.(string); ok {
				message = msg
			} else {
				message = http.StatusText(statusCode)
			}

			_ = response.Base(ctx, statusCode, statusCode, message, nil, nil)
			return
		}

		// Handle other error types (BusinessLogicError, ValidationErrors, etc)
		errHttpCode, errBusinessCode, msg, errCode, errorMap := MapperErrorCode(err)

		// If error map exists, use structured error response
		if errorMap != nil {
			_ = response.ErrorWithMap(ctx, errHttpCode, msg, errorMap)
			return
		}

		_ = response.Base(ctx, errHttpCode, errBusinessCode, msg, nil, errCode)
	}
}
