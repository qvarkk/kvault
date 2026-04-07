package repositories

import (
	"context"
	"database/sql"
	"errors"
	"qvarkk/kvault/internal/domain"

	"github.com/jmoiron/sqlx"
)

type FileRepo struct {
	db *sqlx.DB
}

func NewFileRepo(db *sqlx.DB) *FileRepo {
	return &FileRepo{db: db}
}

const createFileMetaQuery = `
	INSERT INTO files (user_id, original_name, s3_key, size, mime_type, status)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING *
`

const getFileByIDQuery = `
	SELECT * FROM files WHERE id=$1
`

func (r *FileRepo) CreateNew(ctx context.Context, file *domain.File) error {
	return r.db.QueryRowxContext(
		ctx,
		createFileMetaQuery,
		file.UserID,
		file.OriginalName,
		file.S3Key, file.Size,
		file.MimeType,
		file.Status,
	).
		StructScan(file)
}

func (r *FileRepo) GetByID(ctx context.Context, fileID string) (*domain.File, error) {
	var file domain.File
	err := r.db.GetContext(ctx, &file, getFileByIDQuery, fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &file, nil
}
