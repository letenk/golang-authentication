package controller

import (
	"github.com/labstack/echo/v4"
	middleware "github.com/letenk/golang-authentication/middleware/authentication"
)

func (controller *UserController) AddRoutes(e *echo.Echo, authMiddleware middleware.AuthenticationMiddleware) {
	groupV1 := e.Group("/api/v1/user")

	groupV1.GET("/sessions", controller.GetSessions, authMiddleware.Authenticate(true))
	groupV1.DELETE("/sessions/:id", controller.RevokeSession, authMiddleware.Authenticate(true))
}
