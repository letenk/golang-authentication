package db

import (
	"context"

	"github.com/letenk/golang-authentication/bob/models"
	"github.com/stephenafamo/bob"
)

type UserRepository interface {
	Create(ctx context.Context, entity *models.UserSetter) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByPhone(ctx context.Context, phone string) (*models.User, error)
	SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error
	UpdateByID(ctx context.Context, id int64, setter *models.UserSetter) (*models.User, error)
	// Transactional variant
	UpdatePasswordWithExec(ctx context.Context, exec bob.Executor, userID int64, hashedPassword string) error
}
