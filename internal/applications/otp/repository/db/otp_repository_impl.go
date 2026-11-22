package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/domain"
)

type OTPRepositoryImpl struct {
	db *database.Postgres
}

func NewOTPRepository(db *database.Postgres) *OTPRepositoryImpl {
	return &OTPRepositoryImpl{db: db}
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, params CreateOTPParams) (*domain.OTP, error) {
	row := &otpRow{
		UserID:    params.UserID,
		Code:      params.Code,
		Purpose:   string(params.Purpose),
		ExpiresAt: params.ExpiresAt,
	}
	if _, err := r.db.NewInsert(ctx).Model(row).Returning("*").Exec(ctx); err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *OTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*domain.OTP, error) {
	row := new(otpRow)
	err := r.db.NewSelect(ctx).Model(row).
		Where("user_id = ?", userID).
		Where("purpose = ?", string(purpose)).
		Where("code = ?", code).
		Where("expires_at > CURRENT_TIMESTAMP").
		Where("used_at IS NULL").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return nil, translateErr(err)
	}
	return row.toDomain(), nil
}

// MarkUsed only consumes an unused OTP; 0 rows means it was already used.
func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	res, err := r.db.NewUpdate(ctx).Table("otps").
		Set("used_at = CURRENT_TIMESTAMP").
		Where("id = ?", id).
		Where("used_at IS NULL").
		Exec(ctx)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OTPRepositoryImpl) InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error {
	_, err := r.db.NewUpdate(ctx).Table("otps").
		Set("used_at = CURRENT_TIMESTAMP").
		Where("user_id = ?", userID).
		Where("purpose = ?", string(purpose)).
		Where("used_at IS NULL").
		Exec(ctx)
	return err
}

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
