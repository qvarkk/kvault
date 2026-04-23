package repositories

import (
	"context"
	"errors"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var (
	UserFieldID     = "id"
	UserFieldEmail  = "email"
	UserFieldApiKey = "api_key"
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
		Insert("users").Columns("email", "password", "api_key").
		Values(user.Email, user.Password, user.APIKey).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(user)
	return toRepositoryError(err)
}

func (r *UserRepo) IsApiKeyUnique(ctx context.Context, apiKey string) (bool, error) {
	sql, args, err := r.queryBuilder.
		Select("1").From("users").
		Where(sq.Eq{"api_key": apiKey}).
		Limit(1).ToSql()
	if err != nil {
		return false, toRepositoryError(err)
	}

	var exists bool
	err = r.db.Get(&exists, sql, args...)
	err = toRepositoryError(err)
	if errors.Is(err, ErrNotFound) {
		return true, nil
	}

	return false, err
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldID, userID)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldEmail, email)
}

func (r *UserRepo) GetByApiKey(ctx context.Context, apiKey string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldApiKey, apiKey)
}

// Updates API key and returns updated user
func (r *UserRepo) UpdateApiKey(ctx context.Context, userID string, apiKey string) (*domain.User, error) {
	sql, args, err := r.queryBuilder.
		Update("users").Set("api_key", apiKey).Set("updated_at", "now()").
		Where(sq.Eq{"id": userID}).Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var user domain.User
	err = r.db.GetContext(ctx, &user, sql, args...)
	return &user, toRepositoryError(err)
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
