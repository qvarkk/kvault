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

const getUserByEmail = `
SELECT id, email, password, api_key, created_at, updated_at FROM users WHERE email=$1
`

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.GetContext(ctx, &user, getUserByEmail, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
