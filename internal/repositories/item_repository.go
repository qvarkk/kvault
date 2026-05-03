package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

type ItemRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewItemRepo(db *sqlx.DB) *ItemRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &ItemRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *ItemRepo) CreateNew(ctx context.Context, item *domain.Item) error {
	sql, args, err := r.queryBuilder.
		Insert("items").Columns("user_id", "type", "title", "content").
		Values(item.UserID, item.Type, item.Title, item.Content).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(item)
	return toRepositoryError(err)
}

func (r *ItemRepo) List(ctx context.Context, params domain.ListItemFilter) ([]domain.Item, int, error) {
	var items []domain.Item
	var count int

	offset := uint64(params.PageSize * (params.Page - 1))
	baseQuery := r.queryBuilder.
		Select().From("items").Where(sq.Eq{"user_id": params.UserID}).
		Where(sq.Eq{"deleted_at": nil})

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
		if err := r.db.SelectContext(ctx, &items, itemsQuerySql, itemsArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := r.db.GetContext(ctx, &count, countQuerySql, countArgs...); err != nil {
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

func (r *ItemRepo) GetByID(ctx context.Context, itemID string) (*domain.Item, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("items").
		Where(sq.Eq{"id": itemID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var item domain.Item
	err = r.db.GetContext(ctx, &item, sql, args...)
	return &item, toRepositoryError(err)
}

func (r *ItemRepo) GetActiveByIDForUpdate(ctx context.Context, tx *sqlx.Tx, itemID string) (*domain.Item, error) {
	return r.getByIDForUpdate(ctx, tx, itemID, sq.Eq{"deleted_at": nil})
}

func (r *ItemRepo) GetDeletedByIDForUpdate(ctx context.Context, tx *sqlx.Tx, itemID string) (*domain.Item, error) {
	return r.getByIDForUpdate(ctx, tx, itemID, sq.NotEq{"deleted_at": nil})
}

func (r *ItemRepo) getByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID string,
	deletedCondition sq.Sqlizer,
) (*domain.Item, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("items").
		Where(sq.Eq{"id": itemID}).
		Where(deletedCondition).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var item domain.Item
	err = tx.GetContext(ctx, &item, sql, args...)
	return &item, toRepositoryError(err)
}

func (r *ItemRepo) UpdateTx(ctx context.Context, tx *sqlx.Tx, item *domain.Item) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("title", item.Title).
		Set("content", item.Content).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": item.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) SoftDeleteByIDTx(ctx context.Context, tx *sqlx.Tx, itemID string) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("deleted_at", "now()").
		Where(sq.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) RestoreByIDTx(ctx context.Context, tx *sqlx.Tx, itemID string) error {
	sql, args, err := r.queryBuilder.
		Update("items").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) BindTagByItemIDTx(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID, tagID string,
) error {
	sql, args, err := r.queryBuilder.
		Insert("item_tags").
		Columns("item_id", "tag_id", "source").
		Values(itemID, tagID, domain.TagSourceManual).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *ItemRepo) UnbindTagByItemIDTx(
	ctx context.Context,
	tx *sqlx.Tx,
	itemID, tagID string,
) error {
	sql, args, err := r.queryBuilder.
		Delete("item_tags").
		Where(sq.Eq{"item_id": itemID}).
		Where(sq.Eq{"tag_id": tagID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}
