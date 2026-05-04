package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(config Config) (*RedisStore, error) {
	redisOpt := &redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	}

	client := redis.NewClient(redisOpt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &RedisStore{
		client: client,
	}, nil
}

func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) { return nil, nil }

func (s *RedisStore) Set(
	ctx context.Context,
	key string,
	val []byte,
	ttl time.Duration,
) error {
	return nil
}

func (s *RedisStore) Del(ctx context.Context, keys ...string) error { return nil }

func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) { return 0, nil }
