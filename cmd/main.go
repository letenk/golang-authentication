package main

import (
	"context"
	"math/rand/v2"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
)

func main() {
	e := echo.New()

	if err := credential.InitCredentialEnv(); err != nil {
		e.Logger.Fatal(err)
		panic(err)
	}

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
			credential.GetString("application.name"),
			credential.GetString("application.env"),
			credential.GetString("application.port"),
		)

		if err := e.Start(":" + credential.GetString("application.port")); err != nil {
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
