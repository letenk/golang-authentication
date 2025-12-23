package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/letenk/golang-authentication/configs"
	"github.com/letenk/golang-authentication/configs/database"
)

func main() {

	if err := configs.LoadConfig(); err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if err := database.InitDBPostgresSQL(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	defer database.CloseDatabase()

	e := echo.New()

	// Middleware
 	e.Use(middleware.RequestID())
    e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
        Format: "[${time_rfc3339}] ${status} ${method} ${uri} (${latency_human}) ${error}\n",
    }))
	e.Use(middleware.Recover())
	e.Use(middleware.CORS()) 
 	e.Use(middleware.Secure())

  	setupRoutes(e)

   	cfg := configs.GetConfig()

	// Graceful shutdown
	go func() {
  		log.Printf("🚀 %s server started on port %s (env: %s)",
            cfg.App.Name,
            cfg.App.Port,
            cfg.App.Env,
        )

		if err := e.Start(":" + cfg.App.Port); err != nil {
			log.Println("Shutting down the server")
		}
	}()

	gracefulShutdown(e)
}

// setupRoutes setup routing application
func setupRoutes(e *echo.Echo) {
    // Health check endpoint
    e.GET("/health", healthCheckHandler)

    // API v1 group
    v1 := e.Group("/api/v1")

    // Auth routes (akan ditambahkan nanti)
    auth := v1.Group("/auth")
    _ = auth // Avoid unused variable error
    // auth.POST("/register", handler.Register)
    // auth.POST("/login", handler.Login)
}

// healthCheckHandler endpoint for monitoring
func healthCheckHandler(c echo.Context) error {
    // Check database health
    if err := database.HealthCheck(); err != nil {
        return c.JSON(503, map[string]interface{}{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
    }

    cfg := configs.GetConfig()

    return c.JSON(200, map[string]interface{}{
        "status":  "healthy",
        "service": cfg.App.Name,
        "env":     cfg.App.Env,
    })
}

// gracefulShutdown handled graceful shutdown
func gracefulShutdown(e *echo.Echo) {
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down server gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Shutdown Echo server
    if err := e.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }

    log.Println("Server stopped gracefully")
}
