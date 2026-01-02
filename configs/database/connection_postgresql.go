package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/letenk/golang-authentication/configs/credential"
)

var DB *pgxpool.Pool

func InitDBPostgresSQL() error {

	// Build a connection value from environment variable
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		credential.GetString("DB_HOST"),
		credential.GetString("DB_PORT"),
		credential.GetString("DB_USER"),
		credential.GetString("DB_PASSWORD"),
		credential.GetString("DB_NAME"),
		credential.GetString("DB_SSLMODE"),
	)

	// poolConfig for connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("unable to parse database config: %v", &err)
	}

	// Set connection pool settings
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour // Recycle connection each 1 hours
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", &err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	DB = pool
	log.Println("Database connected successfully")
	return nil
}

func CloseDatabase() {
    if DB == nil {
        log.Println("Database already closed or never initialized")
        return
    }

    log.Println("Initiating database connection closure...")

    stat := DB.Stat()
    log.Printf("Connection pool stats before close - Total: %d, Idle: %d, Acquired: %d",
        stat.TotalConns(),
        stat.IdleConns(),
        stat.AcquiredConns(),
    )

    // Close with timeout
    done := make(chan struct{})

    go func() {
        DB.Close()
        close(done)
    }()

    select {
    case <-done:
        log.Println("Database connection closed successfully")
    case <-time.After(5 * time.Second):
        log.Println("Database close timeout after 5 seconds")
        log.Println("Some connections might still be active, forcing shutdown")
    }

    // Set nil for safety
    DB = nil
}

func GetDB() *pgxpool.Pool {
    return DB
}

func HealthCheck() error {
    if DB == nil {
        return fmt.Errorf("database connection is nil")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    return DB.Ping(ctx)
}
