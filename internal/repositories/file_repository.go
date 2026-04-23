package repositories

import (
	"context"
	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
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

func (r *FileRepo) GetByID(ctx context.Context, fileID string) (*domain.File, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("files").Where(sq.Eq{"id": fileID}).ToSql()
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
