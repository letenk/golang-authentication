package database

import (
	"context"

	"github.com/uptrace/bun"
)

// Query constructors for repositories. Each one runs on the transaction carried
// by ctx (see ContextWithTx) or, when there is none, on the pool. Keeping the
// *bun.DB unexported makes these the only way to build a query, so every query
// joins the caller's transaction.

func (db *Postgres) NewSelect(ctx context.Context) *bun.SelectQuery {
	return db.bun.NewSelect().Conn(db.Conn(ctx))
}

func (db *Postgres) NewInsert(ctx context.Context) *bun.InsertQuery {
	return db.bun.NewInsert().Conn(db.Conn(ctx))
}

func (db *Postgres) NewUpdate(ctx context.Context) *bun.UpdateQuery {
	return db.bun.NewUpdate().Conn(db.Conn(ctx))
}

func (db *Postgres) NewDelete(ctx context.Context) *bun.DeleteQuery {
	return db.bun.NewDelete().Conn(db.Conn(ctx))
}

func (db *Postgres) NewRaw(ctx context.Context, query string, args ...any) *bun.RawQuery {
	return db.bun.NewRaw(query, args...).Conn(db.Conn(ctx))
}
