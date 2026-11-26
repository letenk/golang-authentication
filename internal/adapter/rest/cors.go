package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORS lets browsers on allowedOrigins call the API with cookies. An empty
// list falls back to Echo's "*", which browsers never combine with cookies;
// config validation forbids that in production.
func CORS(allowedOrigins []string) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowCredentials: true,
	})
}
