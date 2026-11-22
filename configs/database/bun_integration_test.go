//go:build integration

package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

type testUser struct {
	bun.BaseModel `bun:"table:users"`

	ID        int64  `bun:"id,pk,autoincrement"`
	Password  string `bun:"password"`
	LoginType string `bun:"login_type"`
	CreatedBy int64  `bun:"created_by"`
	UpdatedBy int64  `bun:"updated_by"`
}

func newTestUser() *testUser {
	return &testUser{Password: "x", LoginType: "email", CreatedBy: 1, UpdatedBy: 1}
}

// begin starts a transaction and returns a context carrying it, the same way
// TrxService.WithTx does.
func begin(t *testing.T, db *database.Postgres) (context.Context, *sql.Tx) {
	t.Helper()
	tx, err := db.DB.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })
	return database.ContextWithTx(context.Background(), tx), tx
}

func countUsers(t *testing.T, db *database.Postgres, ctx context.Context) int {
	t.Helper()
	n, err := db.NewSelect(ctx).Model((*testUser)(nil)).Count(ctx)
	require.NoError(t, err)
	return int(n)
}

func TestNewInsert_JoinsTxAndRollsBack(t *testing.T) {
	db := pgtest.Open(t)
	ctxTx, tx := begin(t, db)

	_, err := db.NewInsert(ctxTx).Model(newTestUser()).Exec(ctxTx)
	require.NoError(t, err)

	assert.Equal(t, 1, countUsers(t, db, ctxTx), "NewSelect with a tx ctx must see the uncommitted row")
	assert.Equal(t, 0, countUsers(t, db, context.Background()), "the pool must not see an uncommitted row")

	require.NoError(t, tx.Rollback())
	assert.Equal(t, 0, pgtest.CountRows(t, db.DB, "users"))
}

func TestNewUpdateAndNewDelete_JoinTx(t *testing.T) {
	db := pgtest.Open(t)
	id := pgtest.InsertUser(t, db.DB, "keep@x.com", "")
	ctxTx, tx := begin(t, db)

	_, err := db.NewUpdate(ctxTx).Table("users").Set("password = ?", "y").Where("id = ?", id).Exec(ctxTx)
	require.NoError(t, err)
	_, err = db.NewDelete(ctxTx).Table("users").Where("id = ?", id).Exec(ctxTx)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	var password string
	require.NoError(t, db.DB.QueryRow(`SELECT password FROM users WHERE id = $1`, id).Scan(&password))
	assert.Equal(t, "x", password, "update and delete inside the rolled-back tx must not persist")
}

func TestNewRaw_JoinsTx(t *testing.T) {
	db := pgtest.Open(t)
	ctxTx, tx := begin(t, db)

	_, err := db.NewRaw(ctxTx,
		"INSERT INTO users (password, login_type, created_by, updated_by) VALUES (?, ?, 1, 1)", "x", "email",
	).Exec(ctxTx)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	assert.Equal(t, 0, pgtest.CountRows(t, db.DB, "users"))
}

func TestQueriesWithoutTxUsePool(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()

	_, err := db.NewInsert(ctx).Model(newTestUser()).Exec(ctx)
	require.NoError(t, err)

	assert.Equal(t, 1, pgtest.CountRows(t, db.DB, "users"))
}

// A column added by a migration before the code that uses it is deployed must
// not break existing queries (sqlc expanded SELECT * at generate time; Bun
// scans what the database returns).
func TestUnknownColumnsAreDiscarded(t *testing.T) {
	db := pgtest.Open(t)
	ctx := context.Background()
	id := pgtest.InsertUser(t, db.DB, "wide@x.com", "")

	var row struct {
		ID int64 `bun:"id"`
	}
	err := db.NewRaw(ctx, "SELECT * FROM users WHERE id = ?", id).Scan(ctx, &row)

	require.NoError(t, err)
	assert.Equal(t, id, row.ID)
}
