package transaction

import (
	"context"

	"github.com/stephenafamo/bob"
)

type TrxService interface {
	WithTx(ctx context.Context, fn func(tx bob.Executor) error) error
}
