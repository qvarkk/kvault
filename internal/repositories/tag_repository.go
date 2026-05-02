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

type TagRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewTagRepo(db *sqlx.DB) *TagRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &TagRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *TagRepo) CreateNew(ctx context.Context, tag *domain.Tag) error {
	sql, args, err := r.queryBuilder.
		Insert("tags").
		Columns("user_id", "name").
		Values(tag.UserID, tag.Name).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(tag)
	return toRepositoryError(err)
}

func (r *TagRepo) List(
	ctx context.Context,
	params domain.ListTagFilter,
) ([]domain.Tag, int, error) {
	offset := uint64(params.PageSize * (params.Page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("tags").
		Where(sq.Eq{"user_id": params.UserID})

	if params.Query != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+params.Query+"%")
	}

	tagsSql, tagsArgs, err := baseQuery.
		Columns("*").
		OrderBy(fmt.Sprintf("%s %s", params.Column, params.Direction)).
		Offset(offset).
		Limit(uint64(params.PageSize)).
		ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countSql, countArgs, err := baseQuery.Columns("COUNT(*)").ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	g, _ := errgroup.WithContext(ctx)

	var tags []domain.Tag
	g.Go(func() error {
		if err := r.db.SelectContext(ctx, &tags, tagsSql, tagsArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	var count int
	g.Go(func() error {
		if err := r.db.GetContext(ctx, &count, countSql, countArgs...); err != nil {
			cancel(err)
			return err
		}
		return nil
	})

	_ = g.Wait()

	if cause := context.Cause(ctx); cause != nil {
		return nil, 0, toRepositoryError(err)
	}

	return tags, count, nil
}

func (r *TagRepo) GetByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	tagID string,
) (*domain.Tag, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("tags").
		Where(sq.Eq{"id": tagID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var tag domain.Tag
	err = tx.GetContext(ctx, &tag, sql, args...)
	return &tag, toRepositoryError(err)
}

func (r *TagRepo) UpdateTx(ctx context.Context, tx *sqlx.Tx, tag *domain.Tag) error {
	sql, args, err := r.queryBuilder.
		Update("tags").
		Set("name", tag.Name).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": tag.ID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}
