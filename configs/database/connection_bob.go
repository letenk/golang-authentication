package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/letenk/golang-authentication/configs/credential"
	"github.com/stephenafamo/bob"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type BobDB struct {
	DB   *sql.DB
	Exec bob.Executor
}

func NewSqlBobClient(cfg *credential.DBConnectionConfig) (*BobDB, error) {

	// Build a connection value from environment variable
	dsn := cfg.GetDSNPostgreSQL()

	log.Debug("DSN: ", dsn)

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

	// TODO: toggle app mode and add debug query
	exec := bob.NewDB(db)

	return &BobDB{
		DB:   db,
		Exec: exec,
	}, nil
}

func (db *BobDB) HealthCheck(ctx context.Context) error {
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
