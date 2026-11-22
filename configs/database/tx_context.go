package database

import (
	"context"
	"database/sql"
)

// DBTX is the subset of *sql.DB / *sql.Tx that query code needs. Both values
// satisfy it, and it satisfies bun.IConn, so Bun queries can run on either.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type txContextKey struct{}

// ContextWithTx returns a copy of ctx carrying tx. Repositories called with the
// returned context run their queries inside tx.
func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// TxFromContext returns the transaction stored by ContextWithTx, if any.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx, ok
}

// Conn returns the transaction carried by ctx, or the connection pool when
// ctx carries none.
func (db *Postgres) Conn(ctx context.Context) DBTX {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return db.DB
}
