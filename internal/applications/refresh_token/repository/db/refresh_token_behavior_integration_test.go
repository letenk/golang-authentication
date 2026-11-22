//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	tokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin behaviour the persistence library must preserve; they were
// written against the sqlc implementation before it was replaced.

func TestRefreshTokenRepo_EmptyDeviceFieldsStoredAsEmptyString(t *testing.T) {
	db := pgtest.Open(t)
	repo := tokenRepo.NewRefreshTokenRepository(db)
	userID := pgtest.InsertUser(t, db.DB, "device@x.com", "")

	_, err := repo.Create(context.Background(), tokenRepo.CreateRefreshTokenParams{
		UserID:    userID,
		Token:     "no-device",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)

	var nameNull, idNull, ipNull, uaNull bool
	err = db.DB.QueryRow(
		`SELECT device_name IS NULL, device_id IS NULL, ip_address IS NULL, user_agent IS NULL
		 FROM refresh_tokens WHERE token = $1`, "no-device",
	).Scan(&nameNull, &idNull, &ipNull, &uaNull)
	require.NoError(t, err)
	assert.False(t, nameNull, "device_name must be stored as '', not NULL")
	assert.False(t, idNull, "device_id must be stored as '', not NULL")
	assert.False(t, ipNull, "ip_address must be stored as '', not NULL")
	assert.False(t, uaNull, "user_agent must be stored as '', not NULL")
}

func TestRefreshTokenRepo_RevokeTwiceSucceeds(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.create(t, f.userID, "twice", time.Hour)

	require.NoError(t, f.repo.Revoke(ctx, "twice"))
	assert.NoError(t, f.repo.Revoke(ctx, "twice"), "Revoke is not guarded by revoked_at IS NULL")
}

func TestRefreshTokenRepo_FindActiveEmptyIsNonNil(t *testing.T) {
	f := setup(t)

	got, err := f.repo.FindActiveByUserID(context.Background(), f.userID)

	require.NoError(t, err)
	assert.NotNil(t, got, "sessions must serialise as [] rather than null")
	assert.Len(t, got, 0)
}
