package db

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/stephenafamo/bob"
)

type OTPRepository interface {
	Create(ctx context.Context, params CreateOTPParams) (*models.Otp, error)
	FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*models.Otp, error)
	MarkUsed(ctx context.Context, id int64) error
	InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error
	// Transactional variants — accept an executor from TrxService.WithTx
	CreateWithExec(ctx context.Context, exec bob.Executor, params CreateOTPParams) (*models.Otp, error)
	MarkUsedWithExec(ctx context.Context, exec bob.Executor, id int64) error
	InvalidatePreviousByUserIDAndPurposeWithExec(ctx context.Context, exec bob.Executor, userID int64, purpose OTPPurpose) error
}
