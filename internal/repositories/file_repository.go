package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

type FileRepo struct {
	db           *sqlx.DB
	queryBuilder sq.StatementBuilderType
}

func NewFileRepo(db *sqlx.DB) *FileRepo {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	return &FileRepo{
		db:           db,
		queryBuilder: builder,
	}
}

func (r *FileRepo) CreateNew(ctx context.Context, file *domain.File) error {
	sql, args, err := r.queryBuilder.
		Insert("files").Columns("user_id", "original_name", "s3_key", "size", "mime_type", "status").
		Values(file.UserID, file.OriginalName, file.S3Key, file.Size, file.MimeType, file.Status).
		Suffix("RETURNING *").ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	err = r.db.QueryRowxContext(ctx, sql, args...).StructScan(file)
	return toRepositoryError(err)
}

func (r *FileRepo) List(ctx context.Context, params domain.ListFileFilter) ([]domain.File, int, error) {
	var files []domain.File
	var count int

	offset := uint64(params.PageSize * (params.Page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("files").
		Where(sq.Eq{"user_id": params.UserID}).
		Where(sq.Eq{"deleted_at": nil})

	if params.Query != "" {
		baseQuery = baseQuery.Where("search_vector @@ websearch_to_tsquery('simple', ?)", params.Query)
	}

	if params.MimeType != "" {
		baseQuery = baseQuery.Where(sq.Eq{"mime_type": params.MimeType})
	}

	// TODO: unify orderby with handler somehow, sql injection possible
	// TODO: refactor repetition in ItemsRepo.List
	filesQuery := baseQuery.
		Columns("*").
		OrderBy(fmt.Sprintf("%s %s", params.Column, params.Direction)).
		Offset(offset).
		Limit(uint64(params.PageSize))
	countQuery := baseQuery.Columns("COUNT(*)")

	filesQuerySql, filesArgs, err := filesQuery.ToSql()
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
		if err := r.db.SelectContext(ctx, &files, filesQuerySql, filesArgs...); err != nil {
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

	return files, count, nil
}

func (r *FileRepo) GetByID(ctx context.Context, fileID string) (*domain.File, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("files").Where(sq.Eq{"id": fileID}).
		Where(sq.Eq{"deleted_at": nil}).ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var file domain.File
	err = r.db.GetContext(ctx, &file, sql, args...)
	return &file, toRepositoryError(err)
}

func (r *FileRepo) GetActiveByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	fileID string,
) (*domain.File, error) {
	return r.getByIDForUpdate(ctx, tx, fileID, sq.Eq{"deleted_at": nil})
}

func (r *FileRepo) GetDeletedByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	fileID string,
) (*domain.File, error) {
	return r.getByIDForUpdate(ctx, tx, fileID, sq.NotEq{"deleted_at": nil})
}

func (r *FileRepo) getByIDForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	fileID string,
	deleteCondition sq.Sqlizer,
) (*domain.File, error) {
	sql, args, err := r.queryBuilder.
		Select("*").
		From("files").
		Where(sq.Eq{"id": fileID}).
		Where(deleteCondition).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var file domain.File
	err = tx.GetContext(ctx, &file, sql, args...)
	return &file, toRepositoryError(err)
}

func (r *FileRepo) UpdateTx(
	ctx context.Context,
	tx *sqlx.Tx,
	file *domain.File,
) error {
	sql, args, err := r.queryBuilder.
		Update("files").
		Set("text_content", file.TextContent).
		Set("status", file.Status).
		Set("updated_at", "now()").
		Where(sq.Eq{"id": file.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *FileRepo) SoftDeleteByIDTx(ctx context.Context, tx *sqlx.Tx, fileID string) error {
	sql, args, err := r.queryBuilder.
		Update("files").
		Set("deleted_at", "now()").
		Where(sq.Eq{"id": fileID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *FileRepo) RestoreByIDTx(ctx context.Context, tx *sqlx.Tx, fileID string) error {
	sql, args, err := r.queryBuilder.
		Update("files").
		Set("deleted_at", nil).
		Where(sq.Eq{"id": fileID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}
