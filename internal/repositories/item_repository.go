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
	INSERT INTO items (user_id, type, title, content)
	VALUES ($1, $2, $3, $4)
	RETURNING *
`

const getByIdQuery = `
	SELECT * FROM items WHERE id=$1
`

func (i *ItemRepo) CreateNew(ctx context.Context, item *domain.Item) error {
	err := i.db.QueryRowxContext(ctx, createItemQuery, item.UserID, item.Type, item.Title, item.Content).
		StructScan(item)
	return toRepositoryError(err)
}

func (i *ItemRepo) GetByID(ctx context.Context, itemID string) (*domain.Item, error) {
	var item domain.Item
	err := i.db.GetContext(ctx, &item, getByIdQuery, itemID)
	return &item, toRepositoryError(err)
}
