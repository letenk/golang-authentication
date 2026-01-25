package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/configs/jwt_config"
	customValidator "github.com/letenk/golang-authentication/configs/validator"
	rest "github.com/letenk/golang-authentication/internal/adapter/rest"
)

func main() {
	e := echo.New()

	// Setup custom validator
	e.Validator = customValidator.NewCustomValidator()

	// Setup global HTTP error handler
	customValidator.SetupGlobalHttpUnhandleErrors(e)

	if err := credential.InitCredentialEnv(); err != nil {
		e.Logger.Fatal(err)
		panic(err)
	}

	cfg := credential.Config

	dbConnection, err := database.NewSqlBobClient(&cfg.Database.Configs)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Info("initialized database configuration=", dbConnection)

	jwtConfig, err := jwt_config.NewJWTConfig(&cfg.Auth.JWT)
	if err != nil {
		log.Fatal("Failed to setup JWT config:", err)
	}

	log.Info("Initialized JWT configuration", jwtConfig)

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

	rest.SetupRouteHandler(e, dbConnection, jwtConfig)

	// Graceful shutdown
	go func() {
		log.Printf("🚀 %s server started on port %s (env: %s)",
			cfg.Application.Name,
			cfg.Application.Port,
			cfg.Application.Env,
		)

		if err := e.Start(":" + cfg.Application.Port); err != nil {
			log.Info("shutting down the server")
		}
	}()

	gracefulShutdown(e)
}

// gracefulShutdown handled graceful shutdown
func gracefulShutdown(e *echo.Echo) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
