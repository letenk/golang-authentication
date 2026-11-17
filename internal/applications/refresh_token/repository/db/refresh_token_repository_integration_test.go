//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	tokenRepo "github.com/letenk/golang-authentication/internal/applications/refresh_token/repository/db"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	repo   tokenRepo.RefreshTokenRepository
	userID int64
	other  int64
}

func setup(t *testing.T) fixture {
	t.Helper()
	db := pgtest.Open(t)
	return fixture{
		repo:   tokenRepo.NewRefreshTokenRepository(db),
		userID: pgtest.InsertUser(t, db.DB, "owner@x.com", ""),
		other:  pgtest.InsertUser(t, db.DB, "other@x.com", ""),
	}
}

func (f fixture) create(t *testing.T, userID int64, token string, expiresIn time.Duration) *domain.RefreshToken {
	t.Helper()
	rt, err := f.repo.Create(context.Background(), tokenRepo.CreateRefreshTokenParams{
		UserID:     userID,
		Token:      token,
		DeviceName: "iPhone",
		DeviceID:   "dev-1",
		IPAddress:  "10.0.0.1",
		UserAgent:  "UA/1.0",
		ExpiresAt:  time.Now().UTC().Add(expiresIn),
	})
	require.NoError(t, err)
	return rt
}

func TestRefreshTokenRepo_CreateAndFind(t *testing.T) {
	f := setup(t)
	created := f.create(t, f.userID, "tok-1", time.Hour)

	got, err := f.repo.FindByToken(context.Background(), "tok-1")

	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, f.userID, got.UserID)
	assert.Equal(t, "tok-1", got.Token)
	assert.Equal(t, "iPhone", got.DeviceName)
	assert.Equal(t, "dev-1", got.DeviceID)
	assert.Equal(t, "10.0.0.1", got.IPAddress)
	assert.Equal(t, "UA/1.0", got.UserAgent)
	assert.WithinDuration(t, time.Now().Add(time.Hour), got.ExpiresAt, time.Minute)
	assert.False(t, got.CreatedAt.IsZero())
	assert.Nil(t, got.RevokedAt)
	assert.Equal(t, "", got.ReplacedByToken)
}

func TestRefreshTokenRepo_FindByTokenNotFound(t *testing.T) {
	f := setup(t)

	_, err := f.repo.FindByToken(context.Background(), "missing")

	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestRefreshTokenRepo_FindActiveExcludesRevokedAndExpired(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.create(t, f.userID, "active", time.Hour)
	f.create(t, f.userID, "revoked", time.Hour)
	require.NoError(t, f.repo.Revoke(ctx, "revoked"))
	f.create(t, f.userID, "expired", -time.Hour)
	f.create(t, f.other, "other-user", time.Hour)

	active, err := f.repo.FindActiveByUserID(ctx, f.userID)
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "active", active[0].Token)

	count, err := f.repo.CountActiveByUserID(ctx, f.userID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestRefreshTokenRepo_FindActiveOrderedNewestFirst(t *testing.T) {
	f := setup(t)
	f.create(t, f.userID, "first", time.Hour)
	f.create(t, f.userID, "second", time.Hour)

	active, err := f.repo.FindActiveByUserID(context.Background(), f.userID)

	require.NoError(t, err)
	require.Len(t, active, 2)
	assert.Equal(t, "second", active[0].Token)
	assert.Equal(t, "first", active[1].Token)
}

func TestRefreshTokenRepo_RevokeNotFound(t *testing.T) {
	f := setup(t)

	assert.ErrorIs(t, f.repo.Revoke(context.Background(), "missing"), domain.ErrNotFound)
}

func TestRefreshTokenRepo_ReplaceToken(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.create(t, f.userID, "old", time.Hour)

	require.NoError(t, f.repo.ReplaceToken(ctx, "old", "new"))

	got, err := f.repo.FindByToken(ctx, "old")
	require.NoError(t, err)
	assert.NotNil(t, got.RevokedAt)
	assert.Equal(t, "new", got.ReplacedByToken)
}

func TestRefreshTokenRepo_ReplaceAlreadyRevoked(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.create(t, f.userID, "old", time.Hour)
	require.NoError(t, f.repo.ReplaceToken(ctx, "old", "new-1"))

	err := f.repo.ReplaceToken(ctx, "old", "new-2")

	assert.ErrorIs(t, err, domain.ErrNotFound, "a token already rotated must not be rotated again")
	got, err := f.repo.FindByToken(ctx, "old")
	require.NoError(t, err)
	assert.Equal(t, "new-1", got.ReplacedByToken)
}

func TestRefreshTokenRepo_RevokeAllCountsOnlyActive(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.create(t, f.userID, "a", time.Hour)
	f.create(t, f.userID, "b", time.Hour)
	f.create(t, f.userID, "c", time.Hour)
	require.NoError(t, f.repo.Revoke(ctx, "c"))

	n, err := f.repo.RevokeAllByUserID(ctx, f.userID)

	require.NoError(t, err)
	assert.Equal(t, 2, n)
	count, err := f.repo.CountActiveByUserID(ctx, f.userID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRefreshTokenRepo_RevokeByIDForUser(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	rt := f.create(t, f.userID, "mine", time.Hour)

	assert.ErrorIs(t, f.repo.RevokeByIDForUser(ctx, rt.ID, f.other), domain.ErrNotFound, "another user cannot revoke my session")
	got, err := f.repo.FindByToken(ctx, "mine")
	require.NoError(t, err)
	assert.Nil(t, got.RevokedAt)

	require.NoError(t, f.repo.RevokeByIDForUser(ctx, rt.ID, f.userID))
	assert.ErrorIs(t, f.repo.RevokeByIDForUser(ctx, rt.ID, f.userID), domain.ErrNotFound, "already revoked")
}
