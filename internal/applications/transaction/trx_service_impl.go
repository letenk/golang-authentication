package transaction

import (
	"context"
	"fmt"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stephenafamo/bob"
)

type TrxServiceImpl struct {
	db *database.BobDB
}

func NewTrxService(db *database.BobDB) TrxService {
	return &TrxServiceImpl{db: db}
}

func (r *TrxServiceImpl) WithTx(ctx context.Context, fn func(tx bob.Executor) error) error {
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	txExec := bob.NewTx(tx)

	if err := fn(txExec); err != nil {
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
