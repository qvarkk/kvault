package services

import (
	"context"
	"encoding/json"
	"qvarkk/kvault/logger"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type cachedList[T any] struct {
	Entities []T
	Count    int
}

func listVersion(ctx context.Context, cache CacheStore, versionKey string) int64 {
	val, err := cache.Get(ctx, versionKey)
	if err != nil {
		logger.Logger.Warn("failed to get list cache version",
			zap.String("versionKey", versionKey),
			zap.Error(err),
		)
		return 0
	}
	if val == nil {
		return 0
	}
	version, err := strconv.ParseInt(string(val), 10, 64)
	return version
}

func invalidateListCache(ctx context.Context, cache CacheStore, versionKey string) {
	_, err := cache.Incr(ctx, versionKey)
	if err != nil {
		logger.Logger.Warn("failed to invalidate list cache",
			zap.String("versionKey", versionKey),
			zap.Error(err),
		)
	}
}

func invalidateSingleCache(ctx context.Context, cache CacheStore, key string) {
	if err := cache.Del(ctx, key); err != nil {
		logger.Logger.Warn("failed to invalidate single cache",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}

func getFromCache[T any](ctx context.Context, cache CacheStore, key string) *T {
	cached, err := cache.Get(ctx, key)
	if err != nil || cached == nil {
		return nil
	}
	var result T
	if err := json.Unmarshal(cached, &result); err != nil {
		return nil
	}
	return &result
}

func setToCache[T any](ctx context.Context, cache CacheStore, key string, val T, ttl time.Duration) {
	payload, err := json.Marshal(val)
	if err != nil {
		return
	}
	if err := cache.Set(ctx, key, payload, ttl); err != nil {
		zap.L().Warn("failed to set cache",
			zap.String("key", key),
			zap.Error(err),
		)
	}
}
