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

type OTPRepositoryImpl struct {
	db *database.BobDB
}

func NewOTPRepository(db *database.BobDB) *OTPRepositoryImpl {
	return &OTPRepositoryImpl{db: db}
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, params CreateOTPParams) (*models.Otp, error) {
	return r.CreateWithExec(ctx, r.db.Exec, params)
}

func (r *OTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*models.Otp, error) {
	now := time.Now().UTC()

	otp, err := models.Otps.Query(
		models.SelectWhere.Otps.UserID.EQ(userID),
		models.SelectWhere.Otps.Purpose.EQ(string(purpose)),
		models.SelectWhere.Otps.Code.EQ(code),
		models.SelectWhere.Otps.ExpiresAt.GT(now),
		models.SelectWhere.Otps.UsedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return otp, nil
}

func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	return r.MarkUsedWithExec(ctx, r.db.Exec, id)
}

func (r *OTPRepositoryImpl) InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error {
	return r.InvalidatePreviousByUserIDAndPurposeWithExec(ctx, r.db.Exec, userID, purpose)
}

func (r *OTPRepositoryImpl) CreateWithExec(ctx context.Context, exec bob.Executor, params CreateOTPParams) (*models.Otp, error) {
	setter := &models.OtpSetter{
		UserID:    omit.From(params.UserID),
		Code:      omit.From(params.Code),
		Purpose:   omit.From(string(params.Purpose)),
		ExpiresAt: omit.From(params.ExpiresAt),
		CreatedAt: omitnull.From(time.Now().UTC()),
	}

	result, err := models.Otps.Insert(setter).One(ctx, exec)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *OTPRepositoryImpl) MarkUsedWithExec(ctx context.Context, exec bob.Executor, id int64) error {
	setter := &models.OtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.Otps.Update(
		models.UpdateWhere.Otps.ID.EQ(id),
		setter.UpdateMod(),
	).One(ctx, exec)

	return err
}

func (r *OTPRepositoryImpl) InvalidatePreviousByUserIDAndPurposeWithExec(ctx context.Context, exec bob.Executor, userID int64, purpose OTPPurpose) error {
	setter := &models.OtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.Otps.Update(
		models.UpdateWhere.Otps.UserID.EQ(userID),
		models.UpdateWhere.Otps.Purpose.EQ(string(purpose)),
		models.UpdateWhere.Otps.UsedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, exec)

	return err
}
