package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/domain"
)

type OTPRepositoryImpl struct {
	db *database.BobDB
}

func NewOTPRepository(db *database.BobDB) *OTPRepositoryImpl {
	return &OTPRepositoryImpl{db: db}
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, params CreateOTPParams) (*domain.OTP, error) {
	setter := &models.OtpSetter{
		UserID:    omit.From(params.UserID),
		Code:      omit.From(params.Code),
		Purpose:   omit.From(string(params.Purpose)),
		ExpiresAt: omit.From(params.ExpiresAt),
		CreatedAt: omitnull.From(time.Now().UTC()),
	}

	otp, err := models.Otps.Insert(setter).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, err
	}
	return toDomainOTP(otp), nil
}

func (r *OTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*domain.OTP, error) {
	now := time.Now().UTC()

	otp, err := models.Otps.Query(
		models.SelectWhere.Otps.UserID.EQ(userID),
		models.SelectWhere.Otps.Purpose.EQ(string(purpose)),
		models.SelectWhere.Otps.Code.EQ(code),
		models.SelectWhere.Otps.ExpiresAt.GT(now),
		models.SelectWhere.Otps.UsedAt.IsNull(),
	).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainOTP(otp), nil
}

func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	setter := &models.OtpSetter{
		UsedAt: omitnull.From(time.Now().UTC()),
	}

	_, err := models.Otps.Update(
		models.UpdateWhere.Otps.ID.EQ(id),
		models.UpdateWhere.Otps.UsedAt.IsNull(), // a concurrent request may have consumed it
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))

	return translateErr(err)
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

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func toDomainOTP(o *models.Otp) *domain.OTP {
	return &domain.OTP{
		ID:        o.ID,
		UserID:    o.UserID,
		Code:      o.Code,
		Purpose:   o.Purpose,
		ExpiresAt: o.ExpiresAt,
		UsedAt:    o.UsedAt.Ptr(),
		CreatedAt: o.CreatedAt.GetOrZero(),
	}
}
