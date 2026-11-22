package database_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stretchr/testify/assert"
)

func TestTxFromContext_RoundTrip(t *testing.T) {
	tx := &sql.Tx{}

	got, ok := database.TxFromContext(database.ContextWithTx(context.Background(), tx))

	assert.True(t, ok)
	assert.Same(t, tx, got)
}

func TestTxFromContext_Absent(t *testing.T) {
	got, ok := database.TxFromContext(context.Background())

	assert.False(t, ok)
	assert.Nil(t, got)
}

func TestConn_UsesTxFromContextOtherwiseDB(t *testing.T) {
	sqlDB := &sql.DB{}
	tx := &sql.Tx{}
	db := &database.Postgres{DB: sqlDB}

	assert.Same(t, sqlDB, db.Conn(context.Background()))
	assert.Same(t, tx, db.Conn(database.ContextWithTx(context.Background(), tx)))
}

func TestUniqueViolationConstraint(t *testing.T) {
	unique := &pgconn.PgError{Code: "23505", ConstraintName: "users_email_unique_idx"}

	name, ok := database.UniqueViolationConstraint(fmt.Errorf("insert: %w", unique))
	assert.True(t, ok)
	assert.Equal(t, "users_email_unique_idx", name)

	_, ok = database.UniqueViolationConstraint(&pgconn.PgError{Code: "23503"})
	assert.False(t, ok, "foreign key violation is not a unique violation")

	_, ok = database.UniqueViolationConstraint(errors.New("plain"))
	assert.False(t, ok)

	_, ok = database.UniqueViolationConstraint(nil)
	assert.False(t, ok)
}
