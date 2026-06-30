package services

import (
	"context"
	"io"
	"time"

	"github.com/jmoiron/sqlx"
)

type Transactor interface {
	WithTx(context.Context, func(*sqlx.Tx) error) error
}

type FileStorage interface {
	Upload(ctx context.Context, body io.Reader) (filename string, err error)
	Get(ctx context.Context, filename string) (io.ReadCloser, error)
	Delete(ctx context.Context, filename string) error
	GenerateDownloadUrl(ctx context.Context, filename, originalName string) (url string, expiresAt time.Time, err error)
	GenerateViewUrl(ctx context.Context, filename string) (url string, expiresAt time.Time, err error)
}

type TaskEnqueuer interface {
	EnqueuePdfProcess(ctx context.Context, userID, fileID string) error
}

type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
}
