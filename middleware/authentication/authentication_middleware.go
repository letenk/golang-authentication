package middleware

import "github.com/labstack/echo/v4"

type AuthenticationMiddleware interface {
	Authenticate(required bool) echo.MiddlewareFunc
}