package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/mods"
)

const (
	emailUniqueIndex = "users_email_unique_idx"
	phoneUniqueIndex = "users_phone_unique_idx"
)

type UserRepositoryImpl struct {
	db *database.BobDB
}

func NewUserRepository(db *database.BobDB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) Create(ctx context.Context, params CreateUserParams) (*domain.User, error) {
	now := time.Now().UTC()
	createdBy := params.CreatedBy
	if createdBy == 0 {
		createdBy = SystemUserID
	}

	setter := &models.UserSetter{
		Name:       nullableString(params.Name),
		Email:      nullableString(params.Email),
		Phone:      nullableString(params.Phone),
		Password:   omit.From(params.Password),
		LoginType:  omit.From(params.LoginType),
		IsVerified: omitnull.From(false),
		CreatedBy:  omit.From(createdBy),
		UpdatedBy:  omit.From(createdBy),
		CreatedAt:  omitnull.From(now),
		UpdatedAt:  omitnull.From(now),
	}

	user, err := models.Users.Insert(setter).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainUser(user), nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return r.findOne(ctx, models.SelectWhere.Users.ID.EQ(id))
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, models.SelectWhere.Users.Email.EQ(email))
}

func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return r.findOne(ctx, models.SelectWhere.Users.Phone.EQ(phone))
}

func (r *UserRepositoryImpl) UpdateProfile(ctx context.Context, id int64, params UpdateProfileParams) (*domain.User, error) {
	setter := &models.UserSetter{
		UpdatedBy: omit.From(params.UpdatedBy),
		UpdatedAt: omitnull.From(time.Now().UTC()),
	}
	if params.Name != nil {
		setter.Name = omitnull.From(*params.Name)
	}
	if params.Phone != nil {
		setter.Phone = omitnull.From(*params.Phone)
	}

	return r.updateOne(ctx, id, setter)
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	_, err := r.updateOne(ctx, userID, &models.UserSetter{
		Password:  omit.From(hashedPassword),
		UpdatedBy: omit.From(userID),
		UpdatedAt: omitnull.From(time.Now().UTC()),
	})
	return err
}

func (r *UserRepositoryImpl) MarkEmailVerified(ctx context.Context, userID int64, verifiedAt time.Time) error {
	_, err := r.updateOne(ctx, userID, &models.UserSetter{
		IsVerified: omitnull.From(true),
		VerifiedAt: omitnull.From(verifiedAt),
		UpdatedBy:  omit.From(userID),
		UpdatedAt:  omitnull.From(time.Now().UTC()),
	})
	return err
}

func (r *UserRepositoryImpl) SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error {
	now := time.Now().UTC()
	_, err := r.updateOne(ctx, id, &models.UserSetter{
		DeletedAt: omitnull.From(now),
		DeletedBy: omitnull.From(deletedBy),
		UpdatedAt: omitnull.From(now),
		UpdatedBy: omit.From(deletedBy),
	})
	return err
}

func (r *UserRepositoryImpl) findOne(ctx context.Context, where mods.Where[*dialect.SelectQuery]) (*domain.User, error) {
	user, err := models.Users.Query(
		where,
		models.SelectWhere.Users.DeletedAt.IsNull(),
	).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainUser(user), nil
}

// updateOne applies setter to the live (not soft-deleted) user with id.
func (r *UserRepositoryImpl) updateOne(ctx context.Context, id int64, setter *models.UserSetter) (*domain.User, error) {
	user, err := models.Users.Update(
		models.UpdateWhere.Users.ID.EQ(id),
		models.UpdateWhere.Users.DeletedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainUser(user), nil
}

func nullableString(s string) omitnull.Val[string] {
	if s == "" {
		return omitnull.Val[string]{} // omitted → column default (NULL)
	}
	return omitnull.From(s)
}

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if constraint, ok := database.UniqueViolationConstraint(err); ok {
		switch constraint {
		case emailUniqueIndex:
			return domain.ErrEmailTaken
		case phoneUniqueIndex:
			return domain.ErrPhoneTaken
		}
	}
	return err
}

func toDomainUser(u *models.User) *domain.User {
	return &domain.User{
		ID:         u.ID,
		Name:       u.Name.GetOrZero(),
		Email:      u.Email.GetOrZero(),
		Phone:      u.Phone.GetOrZero(),
		Password:   u.Password,
		LoginType:  u.LoginType,
		IsVerified: u.IsVerified.GetOrZero(),
		VerifiedAt: u.VerifiedAt.Ptr(),
		CreatedAt:  u.CreatedAt.GetOrZero(),
		UpdatedAt:  u.UpdatedAt.GetOrZero(),
	}
}
