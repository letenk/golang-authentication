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


type UserRepositoryImpl struct {
	db *database.BobDB
}

func NewUserRepository(db *database.BobDB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, entity *models.UserSetter) (*models.User, error) {

	entity.CreatedAt = omitnull.From(time.Now().UTC())
	entity.UpdatedAt = omitnull.From(time.Now().UTC())

	if entity.CreatedBy.IsZero() || entity.CreatedBy == omit.From[int64](0) {
		entity.CreatedBy = omit.From(int64(1)) // Default value for system user
	}

	if entity.UpdatedBy.IsZero() || entity.UpdatedBy == omit.From[int64](0) {
		entity.UpdatedBy = omit.From(int64(1)) // Default value for system user
	}

	user, err := models.Users.Insert(entity).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := models.Users.Query(
		models.SelectWhere.Users.Email.EQ(email),
		models.SelectWhere.Users.DeletedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	user, err := models.Users.Query(
		models.SelectWhere.Users.Phone.EQ(phone),
		models.SelectWhere.Users.DeletedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := models.Users.Query(
		models.SelectWhere.Users.ID.EQ(id),
		models.SelectWhere.Users.DeletedAt.IsNull(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) UpdateByID(ctx context.Context, id int64, setter *models.UserSetter) (*models.User, error) {
	setter.UpdatedAt = omitnull.From(time.Now().UTC())

	user, err := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(id),
		models.UpdateWhere.Users.DeletedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) UpdatePasswordWithExec(ctx context.Context, exec bob.Executor, userID int64, hashedPassword string) error {
	setter := &models.UserSetter{
		Password:  omit.From(hashedPassword),
		UpdatedAt: omitnull.From(time.Now().UTC()),
		UpdatedBy: omit.From(userID),
	}

	_, err := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(userID),
		models.UpdateWhere.Users.DeletedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, exec)

	return err
}

func (r *UserRepositoryImpl) SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error {
	now := time.Now().UTC()

	setter := &models.UserSetter{
		DeletedAt: omitnull.From(now),
		DeletedBy: omitnull.From(deletedBy),
		UpdatedAt: omitnull.From(now),
		UpdatedBy: omit.From(deletedBy),
	}

	_, err := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(id),
		models.UpdateWhere.Users.DeletedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, r.db.Exec)

	return err
}
