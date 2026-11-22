package transaction

import "context"

// TrxService runs fn inside a database transaction. fn receives a context that
// carries the transaction; repository calls made with that context join it.
// Calls made with any other context run outside the transaction.
type TrxService interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
