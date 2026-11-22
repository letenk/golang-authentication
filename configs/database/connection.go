package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/letenk/golang-authentication/configs/credential"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Postgres wraps the connection pool. Repositories build queries through the
// New* methods in bun.go, which run on the transaction carried by ctx (started
// by TrxService) or on the pool. The Bun handle is unexported so a repository
// cannot accidentally run a query outside the caller's transaction.
type Postgres struct {
	DB  *sql.DB
	bun *bun.DB
}

// NewPostgresFromDB wraps an open pool. Bun shares the pool; no second
// connection is opened.
func NewPostgresFromDB(sqlDB *sql.DB) *Postgres {
	return &Postgres{
		DB:  sqlDB,
		bun: bun.NewDB(sqlDB, pgdialect.New(), bun.WithDiscardUnknownColumns()),
	}
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

	if err := CheckStandardConformingStrings(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return NewPostgresFromDB(db), nil
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
