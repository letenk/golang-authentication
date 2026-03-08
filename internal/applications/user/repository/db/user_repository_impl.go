package db

import (
	"context"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
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
