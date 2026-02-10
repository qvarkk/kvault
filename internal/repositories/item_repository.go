package repositories

import (
	"context"
	"qvarkk/kvault/internal/domain"

	"github.com/jmoiron/sqlx"
)

type ItemRepo struct {
	db *sqlx.DB
}

func NewItemRepo(db *sqlx.DB) *ItemRepo {
	return &ItemRepo{db: db}
}

const createItemQuery = `
	INSERT INTO items (user_id, type, title, content, file_meta_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING *
`

func (r *ItemRepo) CreateNew(ctx context.Context, item *domain.Item) error {
	return r.db.QueryRowxContext(ctx, createItemQuery, item.UserID, item.Type, item.Title, item.Content, item.FileMetaID).
		StructScan(item)
}
