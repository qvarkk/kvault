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
	INSERT INTO file_meta (s3_key, size, mime_type, status)
	VALUES ($1, $2, $3, $4)
	RETURNING *
`

func (f *FileRepo) CreateNew(ctx context.Context, fileMeta *domain.FileMeta) error {
	return f.db.QueryRowxContext(ctx, createFileMetaQuery, fileMeta.S3Key, fileMeta.Size, fileMeta.MimeType, fileMeta.Status).
		StructScan(fileMeta)
}
