// Package pgtest provides a real PostgreSQL database for integration tests.
//
// Tests using it must carry the `integration` build tag and are run with
// `make test-db-up && make test-integration`. When TEST_DATABASE_URL is not
// set, Open skips the test, so plain `go test ./...` never needs a database.
package pgtest

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/pressly/goose/v3"
	"github.com/stephenafamo/bob"
)

const envDatabaseURL = "TEST_DATABASE_URL"

var (
	migrateOnce sync.Once
	migrateErr  error
)

// Open connects to TEST_DATABASE_URL, applies all migrations once per test
// process, and empties every application table so each test starts clean.
// Identities are restarted so the first inserted user gets id 1, which the
// users.created_by foreign key relies on (id 1 = system user).
func Open(t *testing.T) *database.BobDB {
	t.Helper()

	url := os.Getenv(envDatabaseURL)
	if url == "" {
		t.Skipf("%s is not set; run `make test-db-up` and `make test-integration`", envDatabaseURL)
	}

	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("pgtest: open database: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("pgtest: ping database: %v", err)
	}

	migrateOnce.Do(func() { migrateErr = migrate(sqlDB) })
	if migrateErr != nil {
		t.Fatalf("pgtest: apply migrations: %v", migrateErr)
	}

	if _, err := sqlDB.Exec(`TRUNCATE users, refresh_tokens, otps RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("pgtest: truncate tables: %v", err)
	}

	return &database.BobDB{DB: sqlDB, Exec: bob.NewDB(sqlDB)}
}

// InsertUser inserts a minimal user row with raw SQL and returns its id.
// Empty email or phone is stored as NULL.
func InsertUser(t *testing.T, db *sql.DB, email, phone string) int64 {
	t.Helper()

	var id int64
	err := db.QueryRow(
		`INSERT INTO users (email, phone, password, login_type, created_by, updated_by)
		 VALUES (NULLIF($1, ''), NULLIF($2, ''), 'x', 'email', 1, 1)
		 RETURNING id`,
		email, phone,
	).Scan(&id)
	if err != nil {
		t.Fatalf("pgtest: insert user: %v", err)
	}
	return id
}

// CountRows returns the number of rows in table. table must be a trusted
// identifier (it is interpolated, not parameterised).
func CountRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatalf("pgtest: count rows in %s: %v", table, err)
	}
	return n
}

func migrate(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetLogger(goose.NopLogger())
	return goose.Up(db, migrationsDir())
}

// migrationsDir resolves <repo>/migrations from this file's location so tests
// work regardless of the package directory they run in.
func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}
