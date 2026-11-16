package db

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
)

// OTPRepository methods run inside the caller's transaction when ctx comes from
// TrxService.WithTx.
type OTPRepository interface {
	Create(ctx context.Context, params CreateOTPParams) (*models.Otp, error)
	FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*models.Otp, error)
	MarkUsed(ctx context.Context, id int64) error
	InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error
}
