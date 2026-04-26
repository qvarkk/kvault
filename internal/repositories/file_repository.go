package repositories

import (
	"context"
	"fmt"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

type ListFileParams struct {
	UserID    string
	Query     string
	MimeType  string
	Page      int
	PageSize  int
	Direction string
	Column    string
}

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

func (r *FileRepo) List(ctx context.Context, params ListFileParams) ([]domain.File, int, error) {
	var files []domain.File
	var count int

	offset := uint64(params.PageSize * (params.Page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("files").
		Where(sq.Eq{"user_id": params.UserID})

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
		Select("*").From("files").
		Where(sq.Eq{"id": fileID}).ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var file domain.File
	err = r.db.GetContext(ctx, &file, sql, args...)
	return &file, toRepositoryError(err)
}

// Updates status and returns updated file
func (r *FileRepo) UpdateStatusByID(
	ctx context.Context,
	fileID string,
	status string,
) (*domain.File, error) {
	return r.updateFile(ctx, fileID, map[string]any{
		"status": status,
	})
}

// Updates text content and returns updated file
func (r *FileRepo) UpdateTextContentByID(
	ctx context.Context,
	fileID string,
	textContent string,
) (*domain.File, error) {
	return r.updateFile(ctx, fileID, map[string]any{
		"text_content": textContent,
	})
}

func (r *FileRepo) updateFile(
	ctx context.Context,
	fileID string,
	values map[string]any,
) (*domain.File, error) {
	sql, args, err := r.queryBuilder.
		Update("files").SetMap(values).Set("updated_at", "now()").
		Where(sq.Eq{"id": fileID}).Suffix("RETURNING *").ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var file domain.File
	err = r.db.GetContext(ctx, &file, sql, args...)
	return &file, toRepositoryError(err)
}
