package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/uptrace/bun"
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

func (r *UserRepositoryImpl) Create(ctx context.Context, params CreateUserParams) (*domain.User, error) {
	createdBy := params.CreatedBy
	if createdBy == 0 {
		createdBy = SystemUserID
	}

	row := &userRow{
		Name:       database.NullString(params.Name),
		Email:      database.NullString(params.Email),
		Phone:      database.NullString(params.Phone),
		Password:   params.Password,
		LoginType:  params.LoginType,
		IsVerified: sql.NullBool{Bool: false, Valid: true},
		CreatedBy:  createdBy,
		UpdatedBy:  createdBy,
	}
	if _, err := r.db.NewInsert(ctx).Model(row).Returning("*").Exec(ctx); err != nil {
		return nil, translateErr(err)
	}
	return row.toDomain(), nil
}

func (r *UserRepositoryImpl) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, "email = ?", email)
}

func (r *UserRepositoryImpl) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return r.findOne(ctx, "phone = ?", phone)
}

// findOne returns the single live (not soft-deleted) user matching where.
func (r *UserRepositoryImpl) findOne(ctx context.Context, where string, arg any) (*domain.User, error) {
	row := new(userRow)
	err := r.db.NewSelect(ctx).Model(row).
		Where(where, arg).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, translateErr(err)
	}
	return row.toDomain(), nil
}

// UpdateProfile is a partial update: a nil field is left unchanged, while
// updated_by/updated_at always move.
func (r *UserRepositoryImpl) UpdateProfile(ctx context.Context, id int64, params UpdateProfileParams) (*domain.User, error) {
	row := new(userRow)
	// Model (not Table) so RETURNING scans into row; Set limits the columns written.
	q := audited(r.db.NewUpdate(ctx).Model(row), params.UpdatedBy)
	if params.Name != nil {
		q = q.Set("name = ?", *params.Name)
	}
	if params.Phone != nil {
		q = q.Set("phone = ?", *params.Phone)
	}

	err := q.
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Returning("*").
		Scan(ctx)
	if err != nil {
		return nil, translateErr(err)
	}
	return row.toDomain(), nil
}

func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	return r.updateLive(ctx, userID, r.update(ctx, userID).Set("password = ?", hashedPassword))
}

func (r *UserRepositoryImpl) MarkEmailVerified(ctx context.Context, userID int64, verifiedAt time.Time) error {
	return r.updateLive(ctx, userID, r.update(ctx, userID).
		Set("is_verified = ?", true).
		Set("verified_at = ?", verifiedAt))
}

func (r *UserRepositoryImpl) SoftDeleteByID(ctx context.Context, id int64, deletedBy int64) error {
	return r.updateLive(ctx, id, r.update(ctx, deletedBy).
		Set("deleted_at = CURRENT_TIMESTAMP").
		Set("deleted_by = ?", deletedBy))
}

// update starts an UPDATE on users that records who changed the row and when.
func (r *UserRepositoryImpl) update(ctx context.Context, updatedBy int64) *bun.UpdateQuery {
	return audited(r.db.NewUpdate(ctx).Table("users"), updatedBy)
}

// audited records who changed the row and when.
func audited(q *bun.UpdateQuery, updatedBy int64) *bun.UpdateQuery {
	return q.Set("updated_by = ?", updatedBy).Set("updated_at = CURRENT_TIMESTAMP")
}

// updateLive applies q to the live user id; 0 rows (missing or soft-deleted)
// becomes domain.ErrNotFound.
func (r *UserRepositoryImpl) updateLive(ctx context.Context, id int64, q *bun.UpdateQuery) error {
	res, err := q.Where("id = ?", id).Where("deleted_at IS NULL").Exec(ctx)
	if err != nil {
		return translateErr(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
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
