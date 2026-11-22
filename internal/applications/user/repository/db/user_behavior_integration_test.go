//go:build integration

package db_test

import (
	"context"
	"testing"
	"time"

	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin behaviour the persistence library must preserve; they were
// written against the sqlc implementation before it was replaced.

func TestUserRepo_UpdateProfileNoFields(t *testing.T) {
	db := pgtest.Open(t)
	repo := userRepo.NewUserRepository(db)
	ctx := context.Background()

	u, err := repo.Create(ctx, userRepo.CreateUserParams{Name: "N", Phone: "+628777", Password: "h", LoginType: "phone"})
	require.NoError(t, err)

	readUpdatedAt := func() time.Time {
		var ts time.Time
		require.NoError(t, db.DB.QueryRow(`SELECT updated_at FROM users WHERE id = $1`, u.ID).Scan(&ts))
		return ts
	}
	before := readUpdatedAt()
	time.Sleep(10 * time.Millisecond)

	updated, err := repo.UpdateProfile(ctx, u.ID, userRepo.UpdateProfileParams{UpdatedBy: u.ID})

	require.NoError(t, err)
	assert.Equal(t, "N", updated.Name)
	assert.Equal(t, "+628777", updated.Phone)
	assert.True(t, readUpdatedAt().After(before), "updated_at must move even when no profile field changes")
}

func TestUserRepo_SoftDeletedUserInvisible(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, userRepo.CreateUserParams{Email: "hidden@x.com", Phone: "+628888", Password: "h", LoginType: "email"})
	require.NoError(t, err)
	require.NoError(t, repo.SoftDeleteByID(ctx, u.ID, u.ID))

	_, err = repo.FindByID(ctx, u.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound, "FindByID")
	_, err = repo.FindByPhone(ctx, "+628888")
	assert.ErrorIs(t, err, domain.ErrNotFound, "FindByPhone")
	_, err = repo.UpdateProfile(ctx, u.ID, userRepo.UpdateProfileParams{Name: strPtr("x"), UpdatedBy: u.ID})
	assert.ErrorIs(t, err, domain.ErrNotFound, "UpdateProfile")
	assert.ErrorIs(t, repo.UpdatePassword(ctx, u.ID, "new"), domain.ErrNotFound, "UpdatePassword")
	assert.ErrorIs(t, repo.MarkEmailVerified(ctx, u.ID, time.Now()), domain.ErrNotFound, "MarkEmailVerified")
}
