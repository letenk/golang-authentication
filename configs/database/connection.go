package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/letenk/golang-authentication/configs/credential"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres wraps the connection pool. Repositories reach it through Conn(ctx)
// so they automatically join a transaction started by TrxService.
type Postgres struct {
	DB *sql.DB
}

func NewPostgres(cfg *credential.DBConnectionConfig) (*Postgres, error) {

	// Build a connection value from environment variable
	// Never log the DSN: it contains the database password.
	dsn := cfg.GetDSNPostgreSQL()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed opening connection to DB: %v", err)
	}

	// Set pool
	db.SetMaxIdleConns(cfg.MaxIdleConn)
	db.SetMaxOpenConns(cfg.MaxOpenConn)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &Postgres{DB: db}, nil
}

func (db *Postgres) HealthCheck(ctx context.Context) error {
	if db == nil || db.DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.DB.PingContext(ctx); err != nil {
		return err
	}

	var one int
	if err := db.DB.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		return err
	}

	if one != 1 {
		return fmt.Errorf("databse is unhealthy")
	}

	return nil
}
