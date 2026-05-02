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

func (r *StopwordRepo) GetForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	word, userID string,
) (*domain.Stopword, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("stopwords").
		Where(sq.Eq{"word": word}).
		Where(sq.Eq{"user_id": userID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var stopword domain.Stopword
	err = tx.GetContext(ctx, &stopword, sql, args...)
	return &stopword, toRepositoryError(err)
}

func (r *StopwordRepo) Get(
	ctx context.Context,
	word, userID string,
) (*domain.Stopword, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("stopwords").
		Where(sq.Eq{"word": word}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var stopword domain.Stopword
	err = r.db.GetContext(ctx, &stopword, sql, args...)
	return &stopword, toRepositoryError(err)
}

func (r *StopwordRepo) EnableTx(ctx context.Context, tx *sqlx.Tx, stopword *domain.Stopword) error {
	stopword.IsEnabled = true
	return r.updateIsEnabledTx(ctx, tx, stopword)
}

func (r *StopwordRepo) DisableTx(ctx context.Context, tx *sqlx.Tx, stopword *domain.Stopword) error {
	stopword.IsEnabled = false
	return r.updateIsEnabledTx(ctx, tx, stopword)
}

func (r *StopwordRepo) updateIsEnabledTx(
	ctx context.Context,
	tx *sqlx.Tx,
	stopword *domain.Stopword,
) error {
	sql, args, err := r.queryBuilder.
		Update("stopwords").
		Set("is_enabled", stopword.IsEnabled).
		Set("updated_at", "now()").
		Where(sq.Eq{"word": stopword.Word}).
		Where(sq.Eq{"user_id": stopword.UserID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *StopwordRepo) Delete(
	ctx context.Context,
	stopword *domain.Stopword,
) error {
	sql, args, err := r.queryBuilder.
		Delete("stopwords").
		Where(sq.Eq{"word": stopword.Word}).
		Where(sq.Eq{"user_id": stopword.UserID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}
