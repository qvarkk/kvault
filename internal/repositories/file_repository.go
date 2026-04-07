package repositories

import (
	"context"
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

func (f *FileRepo) CreateNew(ctx context.Context, file *domain.File) error {
	return f.db.QueryRowxContext(
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
