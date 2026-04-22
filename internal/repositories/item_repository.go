package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

type ListItemParams struct {
	UserID    string
	Query     string
	Type      string
	Page      int
	PageSize  int
	Direction string
	Column    string
}

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

func (i *ItemRepo) List(ctx context.Context, params ListItemParams) ([]domain.Item, int, error) {
	var items []domain.Item
	var count int

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	offset := uint64(params.PageSize * (params.Page - 1))
	baseQuery := psql.
		Select().
		From("items").
		Where(sq.Eq{"user_id": params.UserID})

	if params.Query != "" {
		baseQuery = baseQuery.Where("search_vector @@ websearch_to_tsquery('simple', ?)", params.Query)
	}

	if params.Type != "" {
		baseQuery = baseQuery.Where(sq.Eq{"type": params.Type})
	}

	// TODO: unify orderby with handler somehow, sql injection possible
	itemsQuery := baseQuery.
		Columns("*").
		OrderBy(fmt.Sprintf("%s %s", params.Column, params.Direction)).
		Offset(offset).
		Limit(uint64(params.PageSize))
	countQuery := baseQuery.Columns("COUNT(*)")

	itemsQuerySql, itemsArgs, err := itemsQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countQuerySql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := i.db.SelectContext(ctx, &items, itemsQuerySql, itemsArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := i.db.GetContext(ctx, &count, countQuerySql, countArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	_ = g.Wait()

	if cause := context.Cause(ctx); cause != nil {
		return nil, 0, toRepositoryError(cause)
	}

	return items, count, nil
}

func (i *ItemRepo) GetByID(ctx context.Context, itemID string) (*domain.Item, error) {
	var item domain.Item
	err := i.db.GetContext(ctx, &item, getByIdQuery, itemID)
	return &item, toRepositoryError(err)
}
