package db

import (
	"context"
	"time"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/stephenafamo/bob"
)

type EmailVerificationOTPRepository interface {
	Create(ctx context.Context, userID int64, code string, expiresAt time.Time) (*models.EmailVerificationOtp, error)
	FindValidByUserAndCode(ctx context.Context, userID int64, code string) (*models.EmailVerificationOtp, error)
	MarkUsed(ctx context.Context, id int64) error
	InvalidatePreviousByUserID(ctx context.Context, userID int64) error
	// Transactional variants — accept an executor from TrxService.WithTx
	CreateWithExec(ctx context.Context, exec bob.Executor, userID int64, code string, expiresAt time.Time) (*models.EmailVerificationOtp, error)
	MarkUsedWithExec(ctx context.Context, exec bob.Executor, id int64) error
	InvalidatePreviousByUserIDWithExec(ctx context.Context, exec bob.Executor, userID int64) error
}
