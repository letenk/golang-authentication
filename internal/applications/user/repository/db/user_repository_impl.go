package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/db/sqlcgen"
	"github.com/letenk/golang-authentication/internal/domain"
)

const (
	emailUniqueIndex = "users_email_unique_idx"
	phoneUniqueIndex = "users_phone_unique_idx"
)

type UserRepositoryImpl struct {
	db *database.Postgres
}

func NewUserRepository(db *database.Postgres) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// queries runs on the transaction carried by ctx, or on the pool.
func (r *UserRepositoryImpl) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.db.Conn(ctx))
}

func (r *UserRepositoryImpl) Create(ctx context.Context, params CreateUserParams) (*domain.User, error) {
	createdBy := params.CreatedBy
	if createdBy == 0 {
		createdBy = SystemUserID
	}

	user, err := r.queries(ctx).CreateUser(ctx, sqlcgen.CreateUserParams{
		Name:      database.NullString(params.Name),
		Email:     database.NullString(params.Email),
		Phone:     database.NullString(params.Phone),
		Password:  params.Password,
		LoginType: params.LoginType,
		CreatedBy: createdBy,
	})
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainUser(user), nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return one(r.queries(ctx).GetUserByID(ctx, id))
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return one(r.queries(ctx).GetUserByEmail(ctx, database.NullString(email)))
}

func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return one(r.queries(ctx).GetUserByPhone(ctx, database.NullString(phone)))
}

func (r *UserRepositoryImpl) UpdateProfile(ctx context.Context, id int64, params UpdateProfileParams) (*domain.User, error) {
	return one(r.queries(ctx).UpdateUserProfile(ctx, sqlcgen.UpdateUserProfileParams{
		Name:      database.NullStringPtr(params.Name),
		Phone:     database.NullStringPtr(params.Phone),
		UpdatedBy: params.UpdatedBy,
		ID:        id,
	}))
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	return requireRow(r.queries(ctx).UpdateUserPassword(ctx, sqlcgen.UpdateUserPasswordParams{
		Password: hashedPassword,
		ID:       userID,
	}))
}

func (r *UserRepositoryImpl) MarkEmailVerified(ctx context.Context, userID int64, verifiedAt time.Time) error {
	return requireRow(r.queries(ctx).MarkUserEmailVerified(ctx, sqlcgen.MarkUserEmailVerifiedParams{
		VerifiedAt: sql.NullTime{Time: verifiedAt, Valid: true},
		ID:         userID,
	}))
}

func (r *UserRepositoryImpl) SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error {
	return requireRow(r.queries(ctx).SoftDeleteUser(ctx, sqlcgen.SoftDeleteUserParams{
		DeletedBy: sql.NullInt64{Int64: deletedBy, Valid: true},
		ID:        id,
	}))
}

// one maps a single-row query result to the domain, translating errors.
func one(user sqlcgen.User, err error) (*domain.User, error) {
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainUser(user), nil
}

// requireRow turns "no row affected" into domain.ErrNotFound.
func requireRow(affected int64, err error) error {
	if err != nil {
		return translateErr(err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
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

func toDomainUser(u sqlcgen.User) *domain.User {
	return &domain.User{
		ID:         u.ID,
		Name:       u.Name.String,
		Email:      u.Email.String,
		Phone:      u.Phone.String,
		Password:   u.Password,
		LoginType:  u.LoginType,
		IsVerified: u.IsVerified.Bool,
		VerifiedAt: database.TimePtr(u.VerifiedAt),
		CreatedAt:  u.CreatedAt.Time,
		UpdatedAt:  u.UpdatedAt.Time,
	}
}
