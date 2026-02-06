package repositories

import (
	"context"
	"database/sql"
	"errors"
	"qvarkk/kvault/internal/domain"

	"github.com/jmoiron/sqlx"
)

var (
	UserFieldID     = "id"
	UserFieldEmail  = "email"
	UserFieldApiKey = "api_key"
)

const createUserQuery = `
	INSERT INTO users (email, password, api_key)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at
`

const isApiKeyUniqueQuery = `
	SELECT 1 FROM users WHERE api_key=$1 LIMIT 1
`

const getUserByIDQuery = `
	SELECT * FROM users WHERE id=$1
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

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateNew(ctx context.Context, user *domain.User) error {
	return r.db.QueryRowxContext(ctx, createUserQuery, user.Email, user.Password, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) IsApiKeyUnique(ctx context.Context, apiKey string) (bool, error) {
	var exists bool

	err := r.db.Get(&exists, isApiKeyUniqueQuery, apiKey)
	if errors.Is(err, sql.ErrNoRows) {
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
	var user domain.User
	err := r.db.GetContext(ctx, &user, updateApiKeyQuery, apiKey, userID)
	return &user, err
}

func (r *UserRepo) getByField(ctx context.Context, field string, value string) (*domain.User, error) {
	var query string
	switch field {
	case UserFieldID:
		query = getUserByIDQuery
	case UserFieldEmail:
		query = getUserByEmailQuery
	case UserFieldApiKey:
		query = getUserByApiKeyQuery
	}

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}
