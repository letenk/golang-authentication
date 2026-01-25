package controller

import "github.com/labstack/echo/v4"

func (controller *AuthController) AddRoutes(e *echo.Echo) {
	groupV1 := e.Group("/api/v1/auth")

	groupV1.POST("/register", controller.Register)
	// groupV1.POST("/login", controller.Login) // TODO: implement login
}
