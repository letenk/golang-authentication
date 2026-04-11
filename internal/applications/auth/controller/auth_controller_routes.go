package controller

import (
	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	middleware "github.com/letenk/golang-authentication/middleware/authentication"
)

func (controller *AuthController) AddRoutes(e *echo.Echo, authMiddleware middleware.AuthenticationMiddleware) {
	groupV1 := e.Group("/api/v1/auth")

	// Rate limiters: per-IP, in-memory
	// Login: 5 requests/minute (brute force protection)
	loginRateLimiter := echomiddleware.RateLimiter(
		echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(5.0 / 60), Burst: 5},
		),
	)
	// Refresh: 10 requests/minute
	refreshRateLimiter := echomiddleware.RateLimiter(
		echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(10.0 / 60), Burst: 10},
		),
	)

	groupV1.POST("/register", controller.Register)
	groupV1.POST("/login", controller.Login, loginRateLimiter)
	groupV1.POST("/refresh", controller.RefreshToken, refreshRateLimiter)
	groupV1.POST("/logout", controller.Logout)
	groupV1.POST("/logout-all", controller.LogoutAll, authMiddleware.Authenticate(true))
	groupV1.GET("/me", controller.GetMe, authMiddleware.Authenticate(true))
	groupV1.DELETE("/me", controller.DeleteMe, authMiddleware.Authenticate(true))
	groupV1.POST("/forgot-password", controller.ForgotPassword)
	groupV1.POST("/reset-password", controller.ResetPassword)
}
