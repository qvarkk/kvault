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
	Upload(ctx context.Context, key string, body io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	GeneratePresignUrl(ctx context.Context, key, filename string) (url string, expiresAt time.Time, err error)
	GeneratePresignViewUrl(ctx context.Context, key string) (url string, expiresAt time.Time, err error)
}

type TaskEnqueuer interface {
	EnqueuePdfProcess(ctx context.Context, userID, fileID string) error
	EnqueueUrlFetch(ctx context.Context, userID, itemID string) error
}

type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Incr(ctx context.Context, key string) (int64, error)
}
