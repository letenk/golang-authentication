package rest

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	// Registers the generated spec that echoSwagger serves as doc.json.
	_ "github.com/letenk/golang-authentication/api/swagger"
)

// RegisterSwagger serves Swagger UI at /swagger/index.html. In production the
// route is not registered at all, so the UI and the spec return 404.
func RegisterSwagger(e *echo.Echo, env string) {
	if env == "production" {
		return
	}

	e.GET("/swagger/*", echoSwagger.WrapHandler)
}
