//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	otpRepo "github.com/letenk/golang-authentication/internal/applications/otp/repository/db"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (otpRepo.OTPRepository, int64) {
	t.Helper()
	db := pgtest.Open(t)
	return otpRepo.NewOTPRepository(db), pgtest.InsertUser(t, db.DB, "otp@x.com", "")
}

func create(t *testing.T, repo otpRepo.OTPRepository, userID int64, code string, purpose otpRepo.OTPPurpose, expiresIn time.Duration) *domain.OTP {
	t.Helper()
	otp, err := repo.Create(context.Background(), otpRepo.CreateOTPParams{
		UserID:    userID,
		Code:      code,
		Purpose:   purpose,
		ExpiresAt: time.Now().UTC().Add(expiresIn),
	})
	require.NoError(t, err)
	return otp
}

func TestOTPRepo_CreateAndFindValid(t *testing.T) {
	repo, userID := setup(t)
	created := create(t, repo, userID, "12345", otpRepo.OTPPurposePasswordReset, 5*time.Minute)

	got, err := repo.FindValidByUserAndCode(context.Background(), userID, otpRepo.OTPPurposePasswordReset, "12345")

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, "12345", got.Code)
	assert.Equal(t, "password_reset", got.Purpose)
	assert.Nil(t, got.UsedAt)
	assert.False(t, got.CreatedAt.IsZero())
	assert.WithinDuration(t, time.Now().Add(5*time.Minute), got.ExpiresAt, time.Minute)
}

func TestOTPRepo_FindValidIgnoresExpired(t *testing.T) {
	repo, userID := setup(t)
	create(t, repo, userID, "12345", otpRepo.OTPPurposePasswordReset, -time.Minute)

	_, err := repo.FindValidByUserAndCode(context.Background(), userID, otpRepo.OTPPurposePasswordReset, "12345")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOTPRepo_FindValidIsPurposeScoped(t *testing.T) {
	repo, userID := setup(t)
	create(t, repo, userID, "12345", otpRepo.OTPPurposeEmailVerification, 5*time.Minute)

	_, err := repo.FindValidByUserAndCode(context.Background(), userID, otpRepo.OTPPurposePasswordReset, "12345")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOTPRepo_MarkUsedTwice(t *testing.T) {
	repo, userID := setup(t)
	ctx := context.Background()
	otp := create(t, repo, userID, "12345", otpRepo.OTPPurposePasswordReset, 5*time.Minute)

	require.NoError(t, repo.MarkUsed(ctx, otp.ID))
	assert.ErrorIs(t, repo.MarkUsed(ctx, otp.ID), domain.ErrNotFound, "an OTP must not be consumed twice")

	_, err := repo.FindValidByUserAndCode(ctx, userID, otpRepo.OTPPurposePasswordReset, "12345")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOTPRepo_InvalidatePrevious(t *testing.T) {
	repo, userID := setup(t)
	ctx := context.Background()
	create(t, repo, userID, "11111", otpRepo.OTPPurposePasswordReset, 5*time.Minute)
	create(t, repo, userID, "22222", otpRepo.OTPPurposePasswordReset, 5*time.Minute)
	create(t, repo, userID, "33333", otpRepo.OTPPurposeEmailVerification, 5*time.Minute)

	require.NoError(t, repo.InvalidatePreviousByUserIDAndPurpose(ctx, userID, otpRepo.OTPPurposePasswordReset))

	for _, code := range []string{"11111", "22222"} {
		_, err := repo.FindValidByUserAndCode(ctx, userID, otpRepo.OTPPurposePasswordReset, code)
		assert.ErrorIs(t, err, domain.ErrNotFound, code)
	}
	_, err := repo.FindValidByUserAndCode(ctx, userID, otpRepo.OTPPurposeEmailVerification, "33333")
	assert.NoError(t, err, "other purposes are untouched")
}
