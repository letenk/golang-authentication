package rest

import (
	"github.com/labstack/echo/v4"

	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
	jwt_config "github.com/letenk/golang-authentication/configs/jwt_config"
	"github.com/letenk/golang-authentication/internal/applications/auth"
	authController "github.com/letenk/golang-authentication/internal/applications/auth/controller"
	authenticationService "github.com/letenk/golang-authentication/internal/applications/authentication/service"
	userController "github.com/letenk/golang-authentication/internal/applications/user/controller"
	authentication "github.com/letenk/golang-authentication/middleware/authentication"
)

func SetupRouteHandler(e *echo.Echo, db *database.BobDB, jwtConfig *jwt_config.JWTConfig) {

	// TODO: Implement Swagger

	authSvc := authenticationService.NewAuthenticationService(jwtConfig)
	authMiddleware := authentication.NewAuthenticationMiddleware(authSvc)

	authService := auth.InitializeAuthService(db, jwtConfig)
	authCtrl := authController.NewAuthController(authService)
	userCtrl := userController.NewUserController(authService)

	// Setup routes with middleware
	authCtrl.AddRoutes(e, authMiddleware)
	userCtrl.AddRoutes(e, authMiddleware)

	// API v1 group
	v1 := e.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", healthCheckHandler(db))
}

// healthCheckHandler endpoint for monitoring
func healthCheckHandler(db *database.BobDB) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := db.HealthCheck(c.Request().Context()); err != nil {
			return c.JSON(503, map[string]interface{}{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
		}

		return c.JSON(200, map[string]interface{}{
			"status":  "healthy",
			"service": credential.GetString("application.name"),
			"env":     credential.GetString("application.env"),
		})
	}
}
