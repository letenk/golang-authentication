package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/db/sqlcgen"
	"github.com/letenk/golang-authentication/internal/domain"
)

type OTPRepositoryImpl struct {
	db *database.Postgres
}

func NewOTPRepository(db *database.Postgres) *OTPRepositoryImpl {
	return &OTPRepositoryImpl{db: db}
}

// queries runs on the transaction carried by ctx, or on the pool.
func (r *OTPRepositoryImpl) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.db.Conn(ctx))
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, params CreateOTPParams) (*domain.OTP, error) {
	otp, err := r.queries(ctx).CreateOTP(ctx, sqlcgen.CreateOTPParams{
		UserID:    params.UserID,
		Code:      params.Code,
		Purpose:   string(params.Purpose),
		ExpiresAt: params.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return toDomainOTP(otp), nil
}

func (r *OTPRepositoryImpl) FindValidByUserAndCode(ctx context.Context, userID int64, purpose OTPPurpose, code string) (*domain.OTP, error) {
	otp, err := r.queries(ctx).FindValidOTP(ctx, sqlcgen.FindValidOTPParams{
		UserID:  userID,
		Purpose: string(purpose),
		Code:    code,
	})
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainOTP(otp), nil
}

func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id int64) error {
	n, err := r.queries(ctx).MarkOTPUsed(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OTPRepositoryImpl) InvalidatePreviousByUserIDAndPurpose(ctx context.Context, userID int64, purpose OTPPurpose) error {
	return r.queries(ctx).InvalidateOTPsByUserAndPurpose(ctx, sqlcgen.InvalidateOTPsByUserAndPurposeParams{
		UserID:  userID,
		Purpose: string(purpose),
	})
}

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func toDomainOTP(o sqlcgen.Otp) *domain.OTP {
	return &domain.OTP{
		ID:        o.ID,
		UserID:    o.UserID,
		Code:      o.Code,
		Purpose:   o.Purpose,
		ExpiresAt: o.ExpiresAt,
		UsedAt:    database.TimePtr(o.UsedAt),
		CreatedAt: o.CreatedAt.Time, // zero when NULL
	}
}
