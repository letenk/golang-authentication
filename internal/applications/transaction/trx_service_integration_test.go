//go:build integration

package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	otpRepo "github.com/letenk/golang-authentication/internal/applications/otp/repository/db"
	"github.com/letenk/golang-authentication/internal/applications/transaction"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

func setup(t *testing.T) (transaction.TrxService, func(ctx context.Context) error, func() int) {
	t.Helper()
	db := pgtest.Open(t)
	userID := pgtest.InsertUser(t, db.DB, "tx@example.com", "")

	repo := otpRepo.NewOTPRepository(db)
	createOTP := func(ctx context.Context) error {
		_, err := repo.Create(ctx, otpRepo.CreateOTPParams{
			UserID:    userID,
			Code:      "12345",
			Purpose:   otpRepo.OTPPurposePasswordReset,
			ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
		})
		return err
	}
	count := func() int { return pgtest.CountRows(t, db.DB, "otps") }

	return transaction.NewTrxService(db), createOTP, count
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	trx, createOTP, count := setup(t)

	err := trx.WithTx(context.Background(), func(ctx context.Context) error {
		return createOTP(ctx)
	})

	require.NoError(t, err)
	assert.Equal(t, 1, count())
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	trx, createOTP, count := setup(t)

	err := trx.WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, createOTP(ctx))
		return errBoom
	})

	assert.ErrorIs(t, err, errBoom)
	assert.Equal(t, 0, count())
}

func TestWithTx_RollsBackOnPanic(t *testing.T) {
	trx, createOTP, count := setup(t)

	assert.Panics(t, func() {
		_ = trx.WithTx(context.Background(), func(ctx context.Context) error {
			require.NoError(t, createOTP(ctx))
			panic("x")
		})
	})

	assert.Equal(t, 0, count())
}

func TestWithTx_NestedJoinsOuterTx(t *testing.T) {
	trx, createOTP, count := setup(t)

	err := trx.WithTx(context.Background(), func(ctx context.Context) error {
		require.NoError(t, createOTP(ctx))
		require.NoError(t, trx.WithTx(ctx, func(inner context.Context) error {
			return createOTP(inner)
		}))
		return errBoom
	})

	assert.ErrorIs(t, err, errBoom)
	assert.Equal(t, 0, count(), "inner WithTx must join the outer tx and roll back with it")
}

func TestWithTx_WritesOutsideTxCtxNotRolledBack(t *testing.T) {
	trx, createOTP, count := setup(t)
	outer := context.Background()

	err := trx.WithTx(outer, func(_ context.Context) error {
		require.NoError(t, createOTP(outer))
		return errBoom
	})

	assert.ErrorIs(t, err, errBoom)
	assert.Equal(t, 1, count(), "a write made with the non-tx context is not part of the tx")
}
