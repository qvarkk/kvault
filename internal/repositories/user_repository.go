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
		Insert("users").Columns("username", "password").
		Values(user.Username, user.Password).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(user)
	return toRepositoryError(err)
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldID, userID)
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.getByField(ctx, UserFieldUsername, username)
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
