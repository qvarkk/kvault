package redis

import (
	"context"
	"time"
)

type NoopCache struct{}

func NewNoopCache() *NoopCache { return &NoopCache{} }

func (n *NoopCache) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }
func (n *NoopCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	return nil
}
func (n *NoopCache) Del(ctx context.Context, keys ...string) error       { return nil }
func (n *NoopCache) Incr(ctx context.Context, key string) (int64, error) { return 0, nil }
