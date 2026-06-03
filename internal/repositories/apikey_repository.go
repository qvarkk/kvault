package repositories

import (
	"context"
	"qvarkk/kvault/internal/domain"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type ApiKeyRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewApiKeyRepo(db *sqlx.DB) *ApiKeyRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &ApiKeyRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *ApiKeyRepo) Create(ctx context.Context, key *domain.ApiKey) error {
	sql, args, err := r.queryBuilder.
		Insert("api_keys").
		Columns("user_id", "key_hash", "label", "expires_at").
		Values(key.UserID, key.KeyHash, key.Label, key.ExpiresAt).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(key)
	return toRepositoryError(err)
}

func (r *ApiKeyRepo) GetByHash(ctx context.Context, keyHash string) (*domain.ApiKey, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("api_keys").
		Where(sq.Eq{"key_hash": keyHash}).ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var key domain.ApiKey
	err = r.db.GetContext(ctx, &key, sql, args...)
	return &key, toRepositoryError(err)
}

func (r *ApiKeyRepo) IsHashUnique(ctx context.Context, keyHash string) (bool, error) {
	sql, _, err := r.queryBuilder.
		Select("EXISTS(SELECT 1 FROM api_keys WHERE key_hash = ?)").ToSql()
	if err != nil {
		return false, toRepositoryError(err)
	}

	var exists bool
	if err := r.db.GetContext(ctx, &exists, sql, keyHash); err != nil {
		return false, toRepositoryError(err)
	}

	return !exists, nil
}

// Touch slides the key's expiry window forward on use.
func (r *ApiKeyRepo) Touch(ctx context.Context, id string, lastUsed, expires time.Time) error {
	sql, args, err := r.queryBuilder.
		Update("api_keys").
		Set("last_used_at", lastUsed).
		Set("expires_at", expires).
		Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ApiKeyRepo) ListByUser(ctx context.Context, userID string) ([]domain.ApiKey, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("api_keys").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("last_used_at DESC").ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var keys []domain.ApiKey
	err = r.db.SelectContext(ctx, &keys, sql, args...)
	return keys, toRepositoryError(err)
}

func (r *ApiKeyRepo) UpdateLabel(ctx context.Context, id, userID, label string) error {
	sql, args, err := r.queryBuilder.
		Update("api_keys").Set("label", label).
		Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ApiKeyRepo) DeleteByID(ctx context.Context, id, userID string) error {
	sql, args, err := r.queryBuilder.
		Delete("api_keys").
		Where(sq.Eq{"id": id, "user_id": userID}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ApiKeyRepo) DeleteByUserExcept(ctx context.Context, userID, keepID string) error {
	sql, args, err := r.queryBuilder.
		Delete("api_keys").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.NotEq{"id": keepID}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ApiKeyRepo) DeleteExpired(ctx context.Context) error {
	sql, args, err := r.queryBuilder.
		Delete("api_keys").
		Where(sq.Lt{"expires_at": time.Now()}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}
