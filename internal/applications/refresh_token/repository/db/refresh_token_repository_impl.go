package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/internal/domain"
	"github.com/uptrace/bun"
)

type RefreshTokenRepositoryImpl struct {
	db *database.Postgres
}

func NewRefreshTokenRepository(db *database.Postgres) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: db}
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, params CreateRefreshTokenParams) (*domain.RefreshToken, error) {
	row := &refreshTokenRow{
		UserID:     params.UserID,
		Token:      params.Token,
		DeviceName: storedAsGiven(params.DeviceName),
		DeviceID:   storedAsGiven(params.DeviceID),
		IPAddress:  storedAsGiven(params.IPAddress),
		UserAgent:  storedAsGiven(params.UserAgent),
		ExpiresAt:  params.ExpiresAt,
	}
	if _, err := r.db.NewInsert(ctx).Model(row).Returning("*").Exec(ctx); err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *RefreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	row := new(refreshTokenRow)
	if err := r.db.NewSelect(ctx).Model(row).Where("token = ?", token).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return row.toDomain(), nil
}

// active restricts q to sessions that are neither revoked nor expired.
func active(q *bun.SelectQuery, userID int64) *bun.SelectQuery {
	return q.Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > CURRENT_TIMESTAMP")
}

func (r *RefreshTokenRepositoryImpl) FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error) {
	var rows []refreshTokenRow
	if err := active(r.db.NewSelect(ctx).Model(&rows), userID).Order("created_at DESC").Scan(ctx); err != nil {
		return nil, err
	}

	// Non-nil so an empty list serialises as [] rather than null.
	result := make([]*domain.RefreshToken, 0, len(rows))
	for i := range rows {
		result = append(result, rows[i].toDomain())
	}
	return result, nil
}

func (r *RefreshTokenRepositoryImpl) CountActiveByUserID(ctx context.Context, userID int64) (int, error) {
	n, err := active(r.db.NewSelect(ctx).Model((*refreshTokenRow)(nil)), userID).Count(ctx)
	return int(n), err
}

// Revoke is deliberately not limited to unrevoked tokens: revoking twice
// succeeds and refreshes revoked_at, as it always has.
func (r *RefreshTokenRepositoryImpl) Revoke(ctx context.Context, token string) error {
	return requireRow(r.revoke(ctx).Where("token = ?", token).Exec(ctx))
}

// ReplaceToken only rotates an unrevoked token; 0 rows means a concurrent
// request already rotated or revoked it.
func (r *RefreshTokenRepositoryImpl) ReplaceToken(ctx context.Context, oldToken, newToken string) error {
	return requireRow(r.revoke(ctx).
		Set("replaced_by_token = ?", newToken).
		Where("token = ?", oldToken).
		Where("revoked_at IS NULL").
		Exec(ctx))
}

func (r *RefreshTokenRepositoryImpl) RevokeByIDForUser(ctx context.Context, sessionID int64, userID int64) error {
	return requireRow(r.revoke(ctx).
		Where("id = ?", sessionID).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx))
}

func (r *RefreshTokenRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID int64) (int, error) {
	res, err := r.revoke(ctx).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

func (r *RefreshTokenRepositoryImpl) revoke(ctx context.Context) *bun.UpdateQuery {
	return r.db.NewUpdate(ctx).Table("refresh_tokens").Set("revoked_at = CURRENT_TIMESTAMP")
}

// requireRow turns "no row affected" into domain.ErrNotFound.
func requireRow(res sql.Result, err error) error {
	if err != nil {
		return err
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
