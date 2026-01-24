package db

import (
	"context"
	"time"

	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
)

type UserRepositoryImpl struct {
	db *database.BobDB
}

func NewUserRepository(db *database.BobDB) UserRepository {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, entity *models.UserSetter) (*models.User, error) {

	entity.CreatedAt = omitnull.From(time.Now().UTC())
	entity.UpdatedAt = omitnull.From(time.Now().UTC())

	user, err := models.Users.Insert(entity).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := models.Users.Query(
		models.SelectWhere.Users.Email.EQ(email),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*models.User, error) {
	user, err := models.Users.Query(
		models.SelectWhere.Users.Phone.EQ(phone),
	).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}

	return user, nil
}
