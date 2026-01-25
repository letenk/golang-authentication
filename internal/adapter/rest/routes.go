package rest

import (
	"math/rand/v2"
	"strconv"

	"github.com/aarondl/opt/omitnull"
	"github.com/labstack/echo/v4"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/applications/auth"
	authController "github.com/letenk/golang-authentication/internal/applications/auth/controller"
)

func SetupRouteHandler(e *echo.Echo, db *database.BobDB) {

	// TODO: Implement Swagger

	// Initialize controllers using Wire DI
	authService := auth.InitializeAuthService(db)
	authCtrl := authController.NewAuthController(authService)
	// Setup routes
	authCtrl.AddRoutes(e)

	// API v1 group
	v1 := e.Group("/api/v1")

	// Health check endpoint
	v1.GET("/health", healthCheckHandler(db))

	userGroup := v1.Group("/user")

	userGroup.GET("", getUsers(db))
	// userGroup.POST("", createUser(db))
	userGroup.PUT("/:id", updateUser(db))
	userGroup.DELETE("/:id", deleteUser(db))
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

func randomNumber() int {
	return rand.IntN(1000)
}

func updateUser(db *database.BobDB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{
				"message": "invalid user id",
			})
		}

		type request struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		var req request
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{
				"message": "invalid request body",
			})
		}

		setter := &models.UserSetter{
			Name: omitnull.From(req.Name),
			// Email: omit.From(req.Email),
		}

		_, err = models.Users.
			Update(
				models.UpdateWhere.Users.ID.EQ(id),
				setter.UpdateMod(),
			).
			One(ctx, db.Exec)

		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "failed to update user",
				"error":   err.Error(),
			})
		}

		return c.JSON(200, map[string]string{
			"message": "user updated",
		})
	}
}

func getUsers(db *database.BobDB) echo.HandlerFunc {
	return func(c echo.Context) error {

		ctx := c.Request().Context()

		users, err := models.Users.Query().All(ctx, db.Exec)
		if err != nil {
			return c.JSON(500, map[string]interface{}{
				"message": "failed to get users",
				"error":   err.Error(),
			})
		}

		return c.JSON(200, users)
	}
}

func deleteUser(db *database.BobDB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return c.JSON(400, map[string]string{
				"message": "invalid user id",
			})
		}

		_, err = models.Users.
			Delete(
				models.DeleteWhere.Users.ID.EQ(id),
			).
			Exec(ctx, db.Exec)

		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "failed to delete user",
				"error":   err.Error(),
			})
		}

		return c.JSON(200, map[string]string{
			"message": "user deleted",
		})
	}
}
