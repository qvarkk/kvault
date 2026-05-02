package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type StopwordRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewStopwordRepo(db *sqlx.DB) *StopwordRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &StopwordRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *StopwordRepo) CreateNew(ctx context.Context, stopword *domain.Stopword) error {
	sql, args, err := r.queryBuilder.
		Insert("stopwords").
		Columns("user_id", "word", "source", "is_enabled").
		Values(stopword.UserID, stopword.Word, stopword.Source, stopword.IsEnabled).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(stopword)
	return toRepositoryError(err)
}

func (r *StopwordRepo) List(
	ctx context.Context,
	params domain.ListStopwordParams,
) ([]domain.Stopword, error) {
	// TODO: orderby injection
	query := r.queryBuilder.
		Select("*").
		From(fmt.Sprintf("active_stopwords('%s')", params.UserID)).
		OrderBy(fmt.Sprintf("%s %s", params.Column, params.Direction))

	if params.Query != "" {
		query = query.Where("word LIKE ?", "%"+params.Query+"%")
	}
	if params.Source != "" {
		query = query.Where(sq.Eq{"source": params.Source})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var stopwords []domain.Stopword
	if err := r.db.SelectContext(ctx, &stopwords, sql, args...); err != nil {
		fmt.Printf("%s\n", sql)
		return nil, toRepositoryError(err)
	}

	return stopwords, nil
}
