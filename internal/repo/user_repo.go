package repo

import (
	"context"
	"database/sql"
	"errors"
	"qvarkk/kvault/internal/domain"

	"github.com/jmoiron/sqlx"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

const createUserQuery = `
	INSERT INTO users (email, password, api_key)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
`

const isAPIKeyUniqueQuery = `
	SELECT 1 FROM users WHERE api_key=$1 LIMIT 1
`

const getUserByEmailQuery = `
	SELECT * FROM users WHERE email=$1
`

const getUserByApiKeyQuery = `
	SELECT * FROM users WHERE api_key=$1
`

const updateApiKeyQuery = `
	UPDATE users SET api_key=$1 WHERE id=$2 RETURNING *
`

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.QueryRowxContext(ctx, createUserQuery, user.Email, user.Password, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) IsAPIKeyUnique(ctx context.Context, APIKey string) (bool, error) {
	var exists bool

	// TODO: figure a way around repetition of this, or similar to this, fragment of code in every "get" method
	err := r.db.Get(&exists, isAPIKeyUniqueQuery, APIKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.GetContext(ctx, &user, getUserByEmailQuery, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetByApiKey(ctx context.Context, api_key string) (*domain.User, error) {
	var user domain.User

	err := r.db.GetContext(ctx, &user, getUserByApiKeyQuery, api_key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) UpdateApiKey(ctx context.Context, user *domain.User, api_key string) error {
	return r.db.GetContext(ctx, user, updateApiKeyQuery, api_key, user.ID)
}
