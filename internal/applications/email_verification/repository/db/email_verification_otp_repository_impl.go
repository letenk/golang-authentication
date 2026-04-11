package db

import (
	"context"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stephenafamo/bob"
)

type EmailVerificationOTPRepositoryImpl struct {
	db *database.BobDB
}

func NewEmailVerificationOTPRepository(db *database.BobDB) *EmailVerificationOTPRepositoryImpl {
	return &EmailVerificationOTPRepositoryImpl{db: db}
}

func (r *EmailVerificationOTPRepositoryImpl) Create(ctx context.Context, userID int64, code string, expiresAt time.Time) (*models.EmailVerificationOtp, error) {
	return r.CreateWithExec(ctx, r.db.Exec, userID, code, expiresAt)
}

func (r *EmailVerificationOTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, code string) (*models.EmailVerificationOtp, error) {
	now := time.Now().UTC()

	otp, err := models.EmailVerificationOtps.Query(
		models.SelectWhere.EmailVerificationOtps.UserID.EQ(userID),
		models.SelectWhere.EmailVerificationOtps.Code.EQ(code),
		models.SelectWhere.EmailVerificationOtps.ExpiresAt.GT(now),
		models.SelectWhere.EmailVerificationOtps.UsedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *EmailVerificationOTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	return r.MarkUsedWithExec(ctx, r.db.Exec, id)
}

func (r *EmailVerificationOTPRepositoryImpl) InvalidatePreviousByUserID(ctx context.Context, userID int64) error {
	return r.InvalidatePreviousByUserIDWithExec(ctx, r.db.Exec, userID)
}

func (r *EmailVerificationOTPRepositoryImpl) CreateWithExec(ctx context.Context, exec bob.Executor, userID int64, code string, expiresAt time.Time) (*models.EmailVerificationOtp, error) {
	setter := &models.EmailVerificationOtpSetter{
		UserID:    omit.From(userID),
		Code:      omit.From(code),
		ExpiresAt: omit.From(expiresAt),
		CreatedAt: omitnull.From(time.Now().UTC()),
	}

	result, err := models.EmailVerificationOtps.Insert(setter).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *EmailVerificationOTPRepositoryImpl) MarkUsedWithExec(ctx context.Context, exec bob.Executor, id int64) error {
	setter := &models.EmailVerificationOtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.EmailVerificationOtps.Update(
		models.UpdateWhere.EmailVerificationOtps.ID.EQ(id),
		setter.UpdateMod(),
	).One(ctx, exec)

	return err
}

func (r *EmailVerificationOTPRepositoryImpl) InvalidatePreviousByUserIDWithExec(ctx context.Context, exec bob.Executor, userID int64) error {
	setter := &models.EmailVerificationOtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.EmailVerificationOtps.Update(
		models.UpdateWhere.EmailVerificationOtps.UserID.EQ(userID),
		models.UpdateWhere.EmailVerificationOtps.UsedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, exec)

	return err
}
