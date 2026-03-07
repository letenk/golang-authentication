package repository

import (
	"context"
	"time"

	"github.com/aarondl/opt/omitnull"
	"github.com/letenk/golang-authentication/bob/models"
	"github.com/letenk/golang-authentication/configs/database"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

type RefreshTokenRepositoryImpl struct {
	db *database.BobDB
}

func NewRefreshTokenRepository(db *database.BobDB) *RefreshTokenRepositoryImpl {
	return &RefreshTokenRepositoryImpl{db: db}
}

func (r *RefreshTokenRepositoryImpl) Create(ctx context.Context, entity *models.RefreshTokenSetter) (*models.RefreshToken, error) {
	result, err := models.RefreshTokens.Insert(entity).One(ctx, r.db.Exec)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *RefreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	refreshToken, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.Token.EQ(token),
	).One(ctx, r.db.Exec)

	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}

func (r *RefreshTokenRepositoryImpl) FindActiveByUserID(ctx context.Context, userID int64) ([]*models.RefreshToken, error) {
	now := time.Now()

	tokens, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.UserID.EQ(userID),
		models.SelectWhere.RefreshTokens.RevokedAt.IsNull(),
		models.SelectWhere.RefreshTokens.ExpiresAt.GT(now),
		sm.OrderBy(models.RefreshTokens.Columns.CreatedAt).Desc(),
	).All(ctx, r.db.Exec)

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (r *RefreshTokenRepositoryImpl) CountActiveByUserID(ctx context.Context, userID int64) (int, error) {
	now := time.Now()

	count, err := models.RefreshTokens.Query(
		models.SelectWhere.RefreshTokens.UserID.EQ(userID),
		models.SelectWhere.RefreshTokens.RevokedAt.IsNull(),
		models.SelectWhere.RefreshTokens.ExpiresAt.GT(now),
	).Count(ctx, r.db.Exec)

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *RefreshTokenRepositoryImpl) Revoke(ctx context.Context, token string) error {
	now := time.Now()

	setter := &models.RefreshTokenSetter{
		RevokedAt: omitnull.From(now),
	}

	_, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.Token.EQ(token),
		setter.UpdateMod(),
	).One(ctx, r.db.Exec)

	return err
}

func (r *RefreshTokenRepositoryImpl) ReplaceToken(ctx context.Context, oldToken, newToken string) error {
	now := time.Now()

	setter := &models.RefreshTokenSetter{
		RevokedAt:       omitnull.From(now),
		ReplacedByToken: omitnull.From(newToken),
	}

	_, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.Token.EQ(oldToken),
		setter.UpdateMod(),
	).One(ctx, r.db.Exec)

	return err
}

func (r *RefreshTokenRepositoryImpl) RevokeAllByUserID(ctx context.Context, userID int64) (int, error) {
	now := time.Now()

	setter := &models.RefreshTokenSetter{
		RevokedAt: omitnull.From(now),
	}

	result, err := models.RefreshTokens.Update(
		models.UpdateWhere.RefreshTokens.UserID.EQ(userID),
		models.UpdateWhere.RefreshTokens.RevokedAt.IsNull(),
		setter.UpdateMod(),
	).Exec(ctx, r.db.Exec)

	if err != nil {
		return 0, err
	}

	return int(result), nil
}
