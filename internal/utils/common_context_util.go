package utils

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/internal/utils/headers"
)

// GetContextHeaders retrieves header data from echo context
// This function extracts the Header struct that was populated by authentication middleware
func GetContextHeaders(ctx echo.Context, key headers.ContextKey) (*headers.Header, error) {
	value := ctx.Request().Context().Value(key)

	if value == nil {
		return nil, errors.New("value not found in context")
	}

	res, ok := value.(*headers.Header)
	if !ok {
		return nil, errors.New("value in context is not of type *headers.Header")
	}

	return res, nil
}
