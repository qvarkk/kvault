package repo

import (
	"context"
	"database/sql"
	"errors"
	"qvarkk/kvault/internal/domain"

	"github.com/jmoiron/sqlx"
)

var (
	FieldID     = "id"
	FieldEmail  = "email"
	FieldApiKey = "api_key"
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

func (r *UserRepo) IsAPIKeyUnique(ctx context.Context, apiKey string) (bool, error) {
	var exists bool

	// TODO: figure a way around repetition of this, or similar to this, fragment of code in every "get" method
	err := r.db.Get(&exists, isApiKeyUniqueQuery, apiKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	return r.getByField(ctx, FieldID, userID)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getByField(ctx, FieldEmail, email)
}

func (r *UserRepo) GetByApiKey(ctx context.Context, apiKey string) (*domain.User, error) {
	return r.getByField(ctx, FieldApiKey, apiKey)
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
	case FieldID:
		query = getUserByIDQuery
	case FieldEmail:
		query = getUserByEmailQuery
	case FieldApiKey:
		query = getUserByApiKeyQuery
	}

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
