package repositories

import (
	"context"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const (
	UserFieldID       = "id"
	UserFieldUsername = "username"
	UserFieldApiKey   = "api_key_hash"
)

type UserRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &UserRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *UserRepo) CreateNew(ctx context.Context, user *domain.User) error {
	sql, args, err := r.queryBuilder.
		Insert("users").Columns("username", "password", "api_key_hash").
		Values(user.Username, user.Password, user.APIKeyHash).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(user)
	return toRepositoryError(err)
}

func (r *UserRepo) IsApiKeyUnique(ctx context.Context, apiKeyHash string) (bool, error) {
	sql, _, err := r.queryBuilder.
		Select("EXISTS(SELECT 1 FROM users WHERE api_key_hash = ?)").
		ToSql()
	if err != nil {
		return false, toRepositoryError(err)
	}

	var exists bool
	if err := r.db.GetContext(ctx, &exists, sql, apiKeyHash); err != nil {
		return false, toRepositoryError(err)
	}

	return !exists, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldID, userID)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldUsername, username)
}

func (r *UserRepo) GetByApiKeyHash(ctx context.Context, apiKeyHash string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldApiKey, apiKeyHash)
}

// Updates the API key hash and returns the updated user
func (r *UserRepo) UpdateApiKey(ctx context.Context, userID string, apiKeyHash string) (*domain.User, error) {
	sql, args, err := r.queryBuilder.
		Update("users").Set("api_key_hash", apiKeyHash).Set("updated_at", "now()").
		Where(sq.Eq{"id": userID}).Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var user domain.User
	err = r.db.GetContext(ctx, &user, sql, args...)
	return &user, toRepositoryError(err)
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	sql, args, err := r.queryBuilder.
		Update("users").Set("password", passwordHash).Set("updated_at", "now()").
		Where(sq.Eq{"id": userID}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *UserRepo) DeleteByID(ctx context.Context, userID string) error {
	sql, args, err := r.queryBuilder.
		Delete("users").Where(sq.Eq{"id": userID}).ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *UserRepo) getByField(ctx context.Context, field string, value string) (*domain.User, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("users").
		Where(sq.Eq{field: value}).ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var user domain.User
	err = r.db.GetContext(ctx, &user, sql, args...)
	return &user, toRepositoryError(err)
}
