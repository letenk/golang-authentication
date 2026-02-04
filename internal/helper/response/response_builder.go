package response

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type body struct {
	ErrorCode  int                `json:"code"`
	Message    string             `json:"message"`
	Data       interface{}        `json:"data"`
	Error      string             `json:"error,omitempty"`
	Errors     map[string][]string `json:"errors,omitempty"`
	ServerTime string             `json:"serverTime"`
}

func Base(ctx echo.Context, httpCode int, errorCode int, message string, data interface{}, err error) error {

	date := time.Now().Format(time.RFC1123)
	bodyResponse := body{
		ErrorCode:  errorCode,
		Message:    message,
		ServerTime: date,
	}

	if data != nil {
		bodyResponse.Data = data
	}

	if err != nil {
		bodyResponse.Error = err.Error()
	}

	//added header for standard response
	//https://developer.mozilla.org/en-US/docs/Glossary/Response_header
	ctx.Response().Header().Add("date", date)

	return ctx.JSON(httpCode, bodyResponse)
}

func Success(ctx echo.Context, data interface{}) error {
	if data == nil {
		panic(errors.New("success response : data on body is mandatory"))
	}

	return Base(ctx, http.StatusOK, http.StatusOK, http.StatusText(http.StatusOK), data, nil)
}

func SuccessWithMessage(ctx echo.Context, message string, data interface{}) error {
	if data == nil {
		panic(errors.New("success response : data on body is mandatory"))
	}

	return Base(ctx, http.StatusOK, http.StatusOK, message, data, nil)
}

func Error(ctx echo.Context, httpCode int, err error) error {
	return Base(ctx, httpCode, httpCode, http.StatusText(httpCode), nil, err)
}

func ValidationError(ctx echo.Context, errors map[string][]string) error {
	date := time.Now().Format(time.RFC1123)
	bodyResponse := body{
		ErrorCode:  http.StatusBadRequest,
		Message:    "Validation Failed",
		Data:       nil,
		Errors:     errors,
		ServerTime: date,
	}

	ctx.Response().Header().Add("date", date)
	return ctx.JSON(http.StatusBadRequest, bodyResponse)
}

func ErrorWithMap(ctx echo.Context, httpCode int, message string, errors map[string][]string) error {
	date := time.Now().Format(time.RFC1123)
	bodyResponse := body{
		ErrorCode:  httpCode,
		Message:    message,
		Data:       nil,
		Errors:     errors,
		ServerTime: date,
	}

	ctx.Response().Header().Add("date", date)
	return ctx.JSON(httpCode, bodyResponse)
}
