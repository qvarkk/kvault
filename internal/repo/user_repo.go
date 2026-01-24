package repo

import (
	"context"
	"database/sql"
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

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	return r.db.QueryRowxContext(ctx, createUserQuery, user.Email, user.Password, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepo) IsAPIKeyUnique(ctx context.Context, APIKey string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, isAPIKeyUniqueQuery, APIKey)
	if err == sql.ErrNoRows {
		return true, nil
	} else if err != nil {
		return false, err
	} else {
		return false, nil
	}
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.GetContext(ctx, &user, getUserByEmailQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
