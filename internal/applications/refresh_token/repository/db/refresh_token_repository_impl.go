package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/letenk/golang-authentication/configs/database"
	"github.com/letenk/golang-authentication/db/sqlcgen"
	"github.com/letenk/golang-authentication/internal/domain"
)

type RefreshTokenRepositoryImpl struct {
	db *database.BobDB
}

func NewRefreshTokenRepository(db *database.BobDB) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: db}
}

// queries runs on the transaction carried by ctx, or on the pool.
func (r *RefreshTokenRepositoryImpl) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.db.Conn(ctx))
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, params CreateRefreshTokenParams) (*domain.RefreshToken, error) {
	token, err := r.queries(ctx).CreateRefreshToken(ctx, sqlcgen.CreateRefreshTokenParams{
		UserID:     params.UserID,
		Token:      params.Token,
		DeviceName: storedAsGiven(params.DeviceName),
		DeviceID:   storedAsGiven(params.DeviceID),
		IpAddress:  storedAsGiven(params.IPAddress),
		UserAgent:  storedAsGiven(params.UserAgent),
		ExpiresAt:  params.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}
	return toDomainRefreshToken(token), nil
}

func (r *RefreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	rt, err := r.queries(ctx).GetRefreshTokenByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainRefreshToken(rt), nil
}

func (r *RefreshTokenRepositoryImpl) FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error) {
	tokens, err := r.queries(ctx).ListActiveRefreshTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.RefreshToken, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, toDomainRefreshToken(t))
	}
	return result, nil
}

func (r *RefreshTokenRepositoryImpl) CountActiveByUserID(ctx context.Context, userID int64) (int, error) {
	n, err := r.queries(ctx).CountActiveRefreshTokensByUserID(ctx, userID)
	return int(n), err
}

func (r *RefreshTokenRepositoryImpl) Revoke(ctx context.Context, token string) error {
	return requireRow(r.queries(ctx).RevokeRefreshToken(ctx, token))
}

func (r *RefreshTokenRepositoryImpl) ReplaceToken(ctx context.Context, oldToken, newToken string) error {
	return requireRow(r.queries(ctx).ReplaceRefreshToken(ctx, sqlcgen.ReplaceRefreshTokenParams{
		Token:           oldToken,
		ReplacedByToken: sql.NullString{String: newToken, Valid: true},
	}))
}

func (r *RefreshTokenRepositoryImpl) RevokeByIDForUser(ctx context.Context, sessionID int64, userID int64) error {
	return requireRow(r.queries(ctx).RevokeRefreshTokenByIDForUser(ctx, sqlcgen.RevokeRefreshTokenByIDForUserParams{
		ID:     sessionID,
		UserID: userID,
	}))
}

func (r *RefreshTokenRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID int64) (int, error) {
	n, err := r.queries(ctx).RevokeAllRefreshTokensByUserID(ctx, userID)
	return int(n), err
}

// requireRow turns "no row affected" into domain.ErrNotFound.
func requireRow(affected int64, err error) error {
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// storedAsGiven keeps "" as an empty string (not NULL), matching how device
// fields have always been written.
func storedAsGiven(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func toDomainRefreshToken(t sqlcgen.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:              t.ID,
		UserID:          t.UserID,
		Token:           t.Token,
		DeviceName:      t.DeviceName.String,
		DeviceID:        t.DeviceID.String,
		IPAddress:       t.IpAddress.String,
		UserAgent:       t.UserAgent.String,
		ExpiresAt:       t.ExpiresAt,
		CreatedAt:       t.CreatedAt.Time,
		RevokedAt:       database.TimePtr(t.RevokedAt),
		ReplacedByToken: t.ReplacedByToken.String,
	}
}
