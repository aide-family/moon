package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var _ GlobalCache = (*redisGlobalCache)(nil)

type (
	redisGlobalCache struct {
		client *redis.Client
	}
)

func (l *redisGlobalCache) Exists(ctx context.Context, keys ...string) int64 {
	return l.client.Exists(ctx, keys...).Val()
}

func (l *redisGlobalCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) bool {
	return l.client.SetNX(ctx, key, value, ttl).Val()
}

func (l *redisGlobalCache) HGet(ctx context.Context, prefix string, keys string) ([]byte, error) {
	return l.client.HGet(ctx, prefix, keys).Bytes()
}

func (l *redisGlobalCache) HDel(ctx context.Context, prefix string, keys ...string) error {
	return l.client.HDel(ctx, prefix, keys...).Err()
}

func (l *redisGlobalCache) HSet(ctx context.Context, prefix string, values ...[]byte) error {
	args := make([]any, 0, len(values))
	for _, v := range values {
		args = append(args, string(v))
	}

	return l.client.HSet(ctx, prefix, args).Err()
}

func (l *redisGlobalCache) HGetAll(ctx context.Context, prefix string) (map[string][]byte, error) {
	m, err := l.client.HGetAll(ctx, prefix).Result()
	if err != nil {
		return nil, err
	}
	mRes := make(map[string][]byte)
	for k, v := range m {
		mRes[k] = []byte(v)
	}
	return mRes, nil
}

func (l *redisGlobalCache) Get(ctx context.Context, key string) ([]byte, error) {
	return l.client.Get(ctx, key).Bytes()
}

func (l *redisGlobalCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return l.client.Set(ctx, key, value, ttl).Err()
}

func (l *redisGlobalCache) Del(ctx context.Context, keys ...string) error {
	return l.client.Del(ctx, keys...).Err()
}

func (l *redisGlobalCache) Close() error {
	return l.client.Close()
}

func NewRedisGlobalCache(client *redis.Client) GlobalCache {
	return &redisGlobalCache{
		client: client,
	}
}
