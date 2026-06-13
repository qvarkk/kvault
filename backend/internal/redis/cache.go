package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(connConfig ConnConfig, cacheConfig CacheConfig) (*RedisStore, error) {
	redisOpt := &redis.Options{
		Addr:     connConfig.Addr,
		Username: connConfig.Username,
		Password: connConfig.Password,
		DB:       connConfig.DB,
	}

	client := redis.NewClient(redisOpt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &RedisStore{
		client: client,
	}, nil
}

func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get %q: %w", key, err)
	}
	return val, nil
}

func (s *RedisStore) Set(
	ctx context.Context,
	key string,
	val []byte,
	ttl time.Duration,
) error {
	if err := s.client.Set(ctx, key, val, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %q: %w", key, err)
	}
	return nil
}

func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := s.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) {
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis inct %q: %w", key, err)
	}
	return val, nil
}
