//go:build integration

package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/letenk/golang-authentication/internal/applications/transaction"
	userRepo "github.com/letenk/golang-authentication/internal/applications/user/repository/db"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/letenk/golang-authentication/internal/testutil/pgtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRepo(t *testing.T) userRepo.UserRepository {
	t.Helper()
	return userRepo.NewUserRepository(pgtest.Open(t))
}

func emailParams(email string) userRepo.CreateUserParams {
	return userRepo.CreateUserParams{Email: email, Name: "A", Password: "h", LoginType: "email"}
}

func phoneParams(phone string) userRepo.CreateUserParams {
	return userRepo.CreateUserParams{Phone: phone, Password: "h", LoginType: "phone"}
}

func strPtr(s string) *string { return &s }

func TestUserRepo_CreateAndFind(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, emailParams("a@x.com"))
	require.NoError(t, err)

	assert.Equal(t, "a@x.com", created.Email)
	assert.Equal(t, "A", created.Name)
	assert.Equal(t, "", created.Phone)
	assert.Equal(t, "h", created.Password)
	assert.Equal(t, "email", created.LoginType)
	assert.False(t, created.IsVerified)
	assert.Nil(t, created.VerifiedAt)
	assert.False(t, created.CreatedAt.IsZero())

	byEmail, err := repo.FindByEmail(ctx, "a@x.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byEmail.ID)

	byID, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "a@x.com", byID.Email)
}

func TestUserRepo_CreatedByZeroUsesSystemUser(t *testing.T) {
	db := pgtest.Open(t)
	pgtest.InsertUser(t, db.DB, "system@example.com", "") // id 1 = system user
	repo := userRepo.NewUserRepository(db)

	_, err := repo.Create(context.Background(), emailParams("b@x.com")) // CreatedBy == 0

	require.NoError(t, err)
}

func TestUserRepo_TwoPhoneOnlyUsers(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	first, err := repo.Create(ctx, phoneParams("+628111"))
	require.NoError(t, err)
	_, err = repo.Create(ctx, phoneParams("+628222"))
	require.NoError(t, err, "email must be stored as NULL, not '', so phone-only users don't collide")

	found, err := repo.FindByPhone(ctx, "+628111")
	require.NoError(t, err)
	assert.Equal(t, first.ID, found.ID)
	assert.Equal(t, "", found.Email)
}

func TestUserRepo_DuplicateEmail(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	_, err := repo.Create(ctx, emailParams("dup@x.com"))
	require.NoError(t, err)

	_, err = repo.Create(ctx, emailParams("dup@x.com"))

	assert.True(t, errors.Is(err, domain.ErrEmailTaken), "got %v", err)
}

func TestUserRepo_DuplicatePhone(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	_, err := repo.Create(ctx, phoneParams("+628333"))
	require.NoError(t, err)

	_, err = repo.Create(ctx, phoneParams("+628333"))

	assert.True(t, errors.Is(err, domain.ErrPhoneTaken), "got %v", err)
}

func TestUserRepo_SoftDeletedEmailStillTaken(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, emailParams("gone@x.com"))
	require.NoError(t, err)
	require.NoError(t, repo.SoftDeleteByID(ctx, u.ID, u.ID))

	_, err = repo.FindByEmail(ctx, "gone@x.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	_, err = repo.Create(ctx, emailParams("gone@x.com"))
	assert.ErrorIs(t, err, domain.ErrEmailTaken)
}

func TestUserRepo_FindNotFound(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, 999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	_, err = repo.FindByEmail(ctx, "none@x.com")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	_, err = repo.FindByPhone(ctx, "+620000")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUserRepo_UpdateProfilePartial(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, userRepo.CreateUserParams{Email: "p@x.com", Name: "Old", Phone: "+628444", Password: "h", LoginType: "email"})
	require.NoError(t, err)

	updated, err := repo.UpdateProfile(ctx, u.ID, userRepo.UpdateProfileParams{Name: strPtr("New"), UpdatedBy: u.ID})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "+628444", updated.Phone, "nil Phone must leave phone unchanged")

	unchanged, err := repo.UpdateProfile(ctx, u.ID, userRepo.UpdateProfileParams{UpdatedBy: u.ID})
	require.NoError(t, err)
	assert.Equal(t, "New", unchanged.Name)
	assert.Equal(t, "+628444", unchanged.Phone)

	_, err = repo.UpdateProfile(ctx, 999, userRepo.UpdateProfileParams{Name: strPtr("X"), UpdatedBy: 1})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestUserRepo_UpdateProfilePhoneTaken(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	_, err := repo.Create(ctx, phoneParams("+628555"))
	require.NoError(t, err)
	other, err := repo.Create(ctx, phoneParams("+628666"))
	require.NoError(t, err)

	_, err = repo.UpdateProfile(ctx, other.ID, userRepo.UpdateProfileParams{Phone: strPtr("+628555"), UpdatedBy: other.ID})

	assert.ErrorIs(t, err, domain.ErrPhoneTaken)
}

func TestUserRepo_UpdatePasswordAndNotFound(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, emailParams("pw@x.com"))
	require.NoError(t, err)

	require.NoError(t, repo.UpdatePassword(ctx, u.ID, "new-hash"))
	got, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.Equal(t, "new-hash", got.Password)

	assert.ErrorIs(t, repo.UpdatePassword(ctx, 999, "x"), domain.ErrNotFound)
}

func TestUserRepo_MarkEmailVerified(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, emailParams("v@x.com"))
	require.NoError(t, err)

	require.NoError(t, repo.MarkEmailVerified(ctx, u.ID, time.Now().UTC()))

	got, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.True(t, got.IsVerified)
	assert.NotNil(t, got.VerifiedAt)
}

func TestUserRepo_MarkEmailVerifiedInsideRolledBackTx(t *testing.T) {
	db := pgtest.Open(t)
	repo := userRepo.NewUserRepository(db)
	trx := transaction.NewTrxService(db)
	ctx := context.Background()
	u, err := repo.Create(ctx, emailParams("tx@x.com"))
	require.NoError(t, err)

	err = trx.WithTx(ctx, func(txCtx context.Context) error {
		require.NoError(t, repo.MarkEmailVerified(txCtx, u.ID, time.Now().UTC()))
		return errors.New("boom")
	})
	require.Error(t, err)

	got, err := repo.FindByID(ctx, u.ID)
	require.NoError(t, err)
	assert.False(t, got.IsVerified, "update inside a rolled-back tx must not persist")
}

func TestUserRepo_SoftDeleteTwice(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()
	u, err := repo.Create(ctx, emailParams("del@x.com"))
	require.NoError(t, err)

	require.NoError(t, repo.SoftDeleteByID(ctx, u.ID, u.ID))
	assert.ErrorIs(t, repo.SoftDeleteByID(ctx, u.ID, u.ID), domain.ErrNotFound)
}
