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

const CreateItemWithoutFileQuery = `
	INSERT INTO items (type, title, content)
	VALUES ($1, $2, $3)
	RETURNING *
`

const CreateItemWithFileQuery = `
	INSERT INTO items (type, title, content, file_meta_id)
	VALUES ($1, $2, $3, $4)
	RETURNING *
`

func (r *ItemRepo) CreateItem(ctx context.Context, item *domain.Item) error {
	query := CreateItemWithoutFileQuery
	if item.FileMetaID != nil {
		query = CreateItemWithFileQuery
	}

	return r.db.QueryRowxContext(ctx, query, item.Type, item.Title, item.Content, item.FileMetaID).
		StructScan(item)
}
