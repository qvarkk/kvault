package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"

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

func (r *TagRepo) List(ctx context.Context, params domain.ListTagFilter) ([]domain.Tag, int, error) {
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
