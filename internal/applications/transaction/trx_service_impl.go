package transaction

import (
	"context"
	"fmt"

	"github.com/letenk/golang-authentication/configs/database"
)

type TrxServiceImpl struct {
	db *database.BobDB
}

func NewTrxService(db *database.BobDB) TrxService {
	return &TrxServiceImpl{db: db}
}

// WithTx commits when fn returns nil and rolls back when fn returns an error or
// panics (the panic is re-raised). When ctx already carries a transaction, fn
// joins it instead of starting a new one, so the outermost WithTx decides.
func (r *TrxServiceImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := database.TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(database.ContextWithTx(ctx, tx)); err != nil {
		if errRollback := tx.Rollback(); errRollback != nil {
			return fmt.Errorf("%w: rolling back transaction: %v", err, errRollback)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
