package repositories

import (
	"context"
	"strings"
	"unicode"

	"qvarkk/kvault/internal/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

// fileSortColumns whitelists sortable columns for files.
var fileSortColumns = map[string]string{
	"original_name": "original_name",
	"size":          "size",
	"created_at":    "created_at",
}

const fileSortDefault = "created_at"

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

func (r *FileRepo) List(ctx context.Context, f *domain.ListFileFilter) ([]domain.File, int, error) {
	var files []domain.File
	var count int

	page := max(f.Page, 1)
	pageSize := max(f.PageSize, 0)

	//nolint:gosec // G115: page>=1 and pageSize>=0 via max() above
	offset := uint64(pageSize * (page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("files").
		Where(sq.Eq{"user_id": f.UserID}).
		Where(sq.Eq{"deleted_at": nil})

	var tsQuery string
	if f.Query != "" {
		tsQuery = buildFileTsQuery(f.Query)
		if tsQuery != "" {
			baseQuery = baseQuery.Where("search_vector @@ to_tsquery('simple', ?)", tsQuery)
		}
	}

	if f.MimeType != "" {
		baseQuery = baseQuery.Where(sq.Eq{"mime_type": f.MimeType})
	}

	filesQuery := baseQuery.Columns("*")
	if tsQuery != "" {
		filesQuery = filesQuery.OrderByClause(sq.Expr("ts_rank(search_vector, to_tsquery('simple', ?)) DESC", tsQuery))
	}
	//nolint:gosec // G115: pageSize>=0 via max() above
	filesQuery = filesQuery.
		OrderBy(safeOrderBy(f.Column, f.Direction, fileSortColumns, fileSortDefault)).
		Offset(offset).
		Limit(uint64(pageSize))
	countQuery := baseQuery.Columns("COUNT(*)")

	filesQuerySql, filesArgs, err := filesQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countQuerySql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	if err := selectAndCount(ctx, r.db, filesQuerySql, filesArgs, countQuerySql, countArgs, &files, &count); err != nil {
		return nil, 0, err
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

func (r *FileRepo) ListDeleted(ctx context.Context, f *domain.ListFileFilter) ([]domain.File, int, error) {
	var files []domain.File
	var count int

	page := max(f.Page, 1)
	pageSize := max(f.PageSize, 0)

	//nolint:gosec // G115: page>=1 and pageSize>=0 via max() above
	offset := uint64(pageSize * (page - 1))
	baseQuery := r.queryBuilder.
		Select().
		From("files").
		Where(sq.Eq{"user_id": f.UserID}).
		Where(sq.NotEq{"deleted_at": nil})

	countQuery := baseQuery.Columns("COUNT(*)")
	//nolint:gosec // G115: pageSize>=0 via max() above
	filesQuery := baseQuery.Columns("*").
		OrderBy(safeOrderBy(f.Column, f.Direction, fileSortColumns, fileSortDefault)).
		Offset(offset).
		Limit(uint64(pageSize))

	filesQuerySql, filesArgs, err := filesQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	countQuerySql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, toRepositoryError(err)
	}

	if err := selectAndCount(ctx, r.db, filesQuerySql, filesArgs, countQuerySql, countArgs, &files, &count); err != nil {
		return nil, 0, err
	}

	return files, count, nil
}

func (r *FileRepo) GetAllDeleted(ctx context.Context, userID string) ([]domain.File, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("files").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var files []domain.File
	err = r.db.SelectContext(ctx, &files, sql, args...)
	return files, toRepositoryError(err)
}

func (r *FileRepo) PermanentlyDeleteAllDeleted(ctx context.Context, userID string) error {
	sql, args, err := r.queryBuilder.
		Delete("files").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

// HardDeleteByID unconditionally deletes a file row (ignores soft-delete state).
// Used to compensate a failed upload so no orphan record is left behind.
func (r *FileRepo) HardDeleteByID(ctx context.Context, fileID string) error {
	sql, args, err := r.queryBuilder.
		Delete("files").
		Where(sq.Eq{"id": fileID}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *FileRepo) PermanentlyDeleteByIDTx(ctx context.Context, tx *sqlx.Tx, fileID string) error {
	sql, args, err := r.queryBuilder.
		Delete("files").
		Where(sq.Eq{"id": fileID}).
		Where(sq.NotEq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return toRepositoryError(err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	return toRepositoryError(err)
}

func (r *FileRepo) GetAllByUserID(ctx context.Context, userID string) ([]domain.File, error) {
	sql, args, err := r.queryBuilder.
		Select("*").From("files").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, toRepositoryError(err)
	}

	var files []domain.File
	err = r.db.SelectContext(ctx, &files, sql, args...)
	return files, toRepositoryError(err)
}

func buildFileTsQuery(input string) string {
	tokens := strings.Fields(input)
	parts := make([]string, 0, len(tokens))
	for _, t := range tokens {
		var b strings.Builder
		for _, r := range t {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
				b.WriteRune(r)
			}
		}
		if s := b.String(); s != "" {
			parts = append(parts, s+":*")
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " & ")
}
