package controller

import (
	"github.com/labstack/echo/v4"
	middleware "github.com/letenk/golang-authentication/middleware/authentication"
)

func (controller *AuthController) AddRoutes(e *echo.Echo, authMiddleware middleware.AuthenticationMiddleware) {
	groupV1 := e.Group("/api/v1/auth")

	// Public routes (no authentication required)
	groupV1.POST("/register", controller.Register)
	groupV1.POST("/login", controller.Login)
	// groupV1.POST("/refresh-token", controller.RefreshToken)

	// Protected routes (authentication required)
	// groupV1.POST("/logout", controller.Logout, authMiddleware.Authenticate(true))
	// groupV1.POST("/logout-all", controller.LogoutAll, authMiddleware.Authenticate(true))
	// groupV1.GET("/me", controller.GetMe, authMiddleware.Authenticate(true))
}
