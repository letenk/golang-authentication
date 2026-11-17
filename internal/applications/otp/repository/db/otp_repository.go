package db

import (
	"context"

	"github.com/letenk/golang-authentication/internal/domain"
)

// OTPRepository stores one-time codes. Methods return domain.ErrNotFound when
// no row matches or an update affects no row. When ctx comes from
// TrxService.WithTx, methods run inside that transaction.
type OTPRepository interface {
	Create(ctx context.Context, params CreateOTPParams) (*domain.OTP, error)
	// FindValidByUserAndCode returns an unexpired, unused OTP for the purpose.
	FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*domain.OTP, error)
	// MarkUsed consumes an unused OTP; consuming it again returns ErrNotFound.
	MarkUsed(ctx context.Context, id int64) error
	InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error
}
