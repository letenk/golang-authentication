package db

import (
	"context"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
)

type OTPRepositoryImpl struct {
	db *database.BobDB
}

func NewOTPRepository(db *database.BobDB) *OTPRepositoryImpl {
	return &OTPRepositoryImpl{db: db}
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, params CreateOTPParams) (*models.Otp, error) {
	setter := &models.OtpSetter{
		UserID:    omit.From(params.UserID),
		Code:      omit.From(params.Code),
		Purpose:   omit.From(string(params.Purpose)),
		ExpiresAt: omit.From(params.ExpiresAt),
		CreatedAt: omitnull.From(time.Now().UTC()),
	}

	return models.Otps.Insert(setter).One(ctx, r.db.BobExec(ctx))
}

func (r *OTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*models.Otp, error) {
	now := time.Now().UTC()

	return models.Otps.Query(
		models.SelectWhere.Otps.UserID.EQ(userID),
		models.SelectWhere.Otps.Purpose.EQ(string(purpose)),
		models.SelectWhere.Otps.Code.EQ(code),
		models.SelectWhere.Otps.ExpiresAt.GT(now),
		models.SelectWhere.Otps.UsedAt.IsNull(),
	).One(ctx, r.db.BobExec(ctx))
}

func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	setter := &models.OtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.Otps.Update(
		models.UpdateWhere.Otps.ID.EQ(id),
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))

	return err
}

func (r *OTPRepositoryImpl) InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error {
	setter := &models.OtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.Otps.Update(
		models.UpdateWhere.Otps.UserID.EQ(userID),
		models.UpdateWhere.Otps.Purpose.EQ(string(purpose)),
		models.UpdateWhere.Otps.UsedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, r.db.BobExec(ctx))

	return err
}
