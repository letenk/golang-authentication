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

type PasswordResetOTPRepositoryImpl struct {
	db *database.BobDB
}

func NewPasswordResetOTPRepository(db *database.BobDB) *PasswordResetOTPRepositoryImpl {
	return &PasswordResetOTPRepositoryImpl{db: db}
}

func (r *PasswordResetOTPRepositoryImpl) Create(ctx context.Context, userID int64, code string, expiresAt time.Time) (*models.PasswordResetOtp, error) {
	return r.CreateWithExec(ctx, r.db.Exec, userID, code, expiresAt)
}

func (r *PasswordResetOTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, code string) (*models.PasswordResetOtp, error) {
	now := time.Now().UTC()

	otp, err := models.PasswordResetOtps.Query(
		models.SelectWhere.PasswordResetOtps.UserID.EQ(userID),
		models.SelectWhere.PasswordResetOtps.Code.EQ(code),
		models.SelectWhere.PasswordResetOtps.ExpiresAt.GT(now),
		models.SelectWhere.PasswordResetOtps.UsedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *PasswordResetOTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	return r.MarkUsedWithExec(ctx, r.db.Exec, id)
}

func (r *PasswordResetOTPRepositoryImpl) InvalidatePreviousByUserID(ctx context.Context, userID int64) error {
	return r.InvalidatePreviousByUserIDWithExec(ctx, r.db.Exec, userID)
}

func (r *PasswordResetOTPRepositoryImpl) CreateWithExec(ctx context.Context, exec bob.Executor, userID int64, code string, expiresAt time.Time) (*models.PasswordResetOtp, error) {
	setter := &models.PasswordResetOtpSetter{
		UserID:    omit.From(userID),
		Code:      omit.From(code),
		ExpiresAt: omit.From(expiresAt),
		CreatedAt: omitnull.From(time.Now().UTC()),
	}

	result, err := models.PasswordResetOtps.Insert(setter).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *PasswordResetOTPRepositoryImpl) MarkUsedWithExec(ctx context.Context, exec bob.Executor, id int64) error {
	setter := &models.PasswordResetOtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.PasswordResetOtps.Update(
		models.UpdateWhere.PasswordResetOtps.ID.EQ(id),
		setter.UpdateMod(),
	).One(ctx, exec)

	return err
}

func (r *PasswordResetOTPRepositoryImpl) InvalidatePreviousByUserIDWithExec(ctx context.Context, exec bob.Executor, userID int64) error {
	setter := &models.PasswordResetOtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.PasswordResetOtps.Update(
		models.UpdateWhere.PasswordResetOtps.UserID.EQ(userID),
		models.UpdateWhere.PasswordResetOtps.UsedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, exec)

	return err
}
