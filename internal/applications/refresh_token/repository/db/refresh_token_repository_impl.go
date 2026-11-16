package repository

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
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type RefreshTokenRepositoryImpl struct {
	db *database.BobDB
}

func NewRefreshTokenRepository(db *database.BobDB) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: db}
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, params CreateRefreshTokenParams) (*domain.RefreshToken, error) {
	setter := &models.RefreshTokenSetter{
		UserID:     omit.From(params.UserID),
		Token:      omit.From(params.Token),
		DeviceName: omitnull.From(params.DeviceName),
		DeviceID:   omitnull.From(params.DeviceID),
		IPAddress:  omitnull.From(params.IPAddress),
		UserAgent:  omitnull.From(params.UserAgent),
		ExpiresAt:  omit.From(params.ExpiresAt),
	}

	token, err := models.RefreshTokens.Insert(setter).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, err
	}
	return toDomainRefreshToken(token), nil
}

func (r *RefreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	rt, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.Token.EQ(token),
	).One(ctx, r.db.BobExec(ctx))
	if err != nil {
		return nil, translateErr(err)
	}
	return toDomainRefreshToken(rt), nil
}

func (r *RefreshTokenRepositoryImpl) FindActiveByUserID(ctx context.Context, userID int64) ([]*domain.RefreshToken, error) {
	tokens, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.UserID.EQ(userID),
		models.SelectWhere.RefreshTokens.RevokedAt.IsNull(),
		models.SelectWhere.RefreshTokens.ExpiresAt.GT(time.Now().UTC()),
		sm.OrderBy(models.RefreshTokens.Columns.CreatedAt).Desc(),
	).All(ctx, r.db.BobExec(ctx))
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
	count, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.UserID.EQ(userID),
		models.SelectWhere.RefreshTokens.RevokedAt.IsNull(),
		models.SelectWhere.RefreshTokens.ExpiresAt.GT(time.Now().UTC()),
	).Count(ctx, r.db.BobExec(ctx))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *RefreshTokenRepositoryImpl) Revoke(ctx context.Context, token string) error {
	setter := &models.RefreshTokenSetter{RevokedAt: omitnull.From(time.Now().UTC())}

	_, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.Token.EQ(token),
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))
	return translateErr(err)
}

func (r *RefreshTokenRepositoryImpl) ReplaceToken(ctx context.Context, oldToken, newToken string) error {
	setter := &models.RefreshTokenSetter{
		RevokedAt:       omitnull.From(time.Now().UTC()),
		ReplacedByToken: omitnull.From(newToken),
	}

	_, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.Token.EQ(oldToken),
		models.UpdateWhere.RefreshTokens.RevokedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))
	return translateErr(err)
}

func (r *RefreshTokenRepositoryImpl) RevokeByIDForUser(ctx context.Context, sessionID int64, userID int64) error {
	setter := &models.RefreshTokenSetter{RevokedAt: omitnull.From(time.Now().UTC())}

	_, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.ID.EQ(sessionID),
		models.UpdateWhere.RefreshTokens.UserID.EQ(userID),
		models.UpdateWhere.RefreshTokens.RevokedAt.IsNull(),
		setter.UpdateMod(),
	).One(ctx, r.db.BobExec(ctx))
	return translateErr(err)
}

func (r *RefreshTokenRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID int64) (int, error) {
	setter := &models.RefreshTokenSetter{RevokedAt: omitnull.From(time.Now().UTC())}

	n, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.UserID.EQ(userID),
		models.UpdateWhere.RefreshTokens.RevokedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, r.db.BobExec(ctx))
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func translateErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}

func toDomainRefreshToken(t *models.RefreshToken) *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:              t.ID,
		UserID:          t.UserID,
		Token:           t.Token,
		DeviceName:      t.DeviceName.GetOrZero(),
		DeviceID:        t.DeviceID.GetOrZero(),
		IPAddress:       t.IPAddress.GetOrZero(),
		UserAgent:       t.UserAgent.GetOrZero(),
		ExpiresAt:       t.ExpiresAt,
		CreatedAt:       t.CreatedAt.GetOrZero(),
		RevokedAt:       t.RevokedAt.Ptr(),
		ReplacedByToken: t.ReplacedByToken.GetOrZero(),
	}
}
