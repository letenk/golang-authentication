package controller

import (
	"github.com/labstack/echo/v4"
	middleware "github.com/letenk/golang-authentication/middleware/authentication"
)

func (controller *AuthController) AddRoutes(e *echo.Echo, authMiddleware middleware.AuthenticationMiddleware) {
	groupV1 := e.Group("/api/v1/auth")

	groupV1.POST("/register", controller.Register)
	groupV1.POST("/login", controller.Login)
}
