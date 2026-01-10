package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
)

func main() {
	e := echo.New()

	if err := credential.InitCredentialEnv(); err != nil {
		e.Logger.Fatal(err)
		panic(err)
	}

	cfg := credential.Config

	dbConnection, err := database.NewSqlBobClient()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Info("initialized database configuration=", dbConnection)

	defer func(dbConnection *database.BobDB) {
		err := dbConnection.DB.Close()
		if err != nil {
			log.Fatalf("error initialized database configuration=", err)
		}
	}(dbConnection)

	// Middleware
	e.Use(middleware.RequestID())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "[${time_rfc3339}] ${status} ${method} ${uri} (${latency_human}) ${error}\n",
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())

	setupRoutes(e, dbConnection)

	// Graceful shutdown
	go func() {
		log.Printf("🚀 %s server started on port %s (env: %s)",
			cfg.Application.Name,
			cfg.Application.Env,
			cfg.Application.Port,
		)

		if err := e.Start(":" + cfg.Application.Port); err != nil {
			log.Info("Shutting down the server")
		}
	}()

	gracefulShutdown(e)
}

// setupRoutes setup routing application
func setupRoutes(e *echo.Echo, db *database.BobDB) {
	// Health check endpoint
	e.GET("/health", healthCheckHandler(db))

	// API v1 group
	v1 := e.Group("/api/v1")

	userGroup := v1.Group("/user")
	userGroup.GET("", getUsers(db))
	userGroup.POST("", createUser(db))
	userGroup.PUT("/:id", updateUser(db))
	userGroup.DELETE("/:id", deleteUser(db))

	// Auth routes (akan ditambahkan nanti)
	auth := v1.Group("/auth")
	_ = auth // Avoid unused variable error
	// auth.POST("/register", handler.Register)
	// auth.POST("/login", handler.Login)
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

func createUser(db *database.BobDB) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		user := models.UserSetter{
			Version: omit.From(int64(randomNumber())),

			CreatedAt: omitnull.From(time.Now()),
			CreatedBy: omit.From(int64(1)),

			UpdatedAt: omitnull.From(time.Now()),
			UpdatedBy: omit.From(int64(1)),

			DeletedAt: omitnull.From(time.Now()),

			Name:  omitnull.From(fmt.Sprintf("name%d", randomNumber())),
			Email: omit.From(fmt.Sprintf("user%d@mail.com", randomNumber())),
			Phone: omitnull.From(fmt.Sprintf("0812%d", randomNumber())),

			Password:  omit.From("hashed-password"), // anggap sudah di-hash
			LoginType: omit.From("email"),
		}

		_, err := models.Users.Insert(&user).Exec(ctx, db.Exec)
		if err != nil {
			return c.JSON(500, map[string]string{
				"message": "failed to create user",
				"error":   err.Error(),
			})
		}

		return c.JSON(201, user)
	}
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
			Name:  omitnull.From(req.Name),
			Email: omit.From(req.Email),
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

// gracefulShutdown handled graceful shutdown
func gracefulShutdown(e *echo.Echo) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Echo server
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Info("Server stopped gracefully")
}
