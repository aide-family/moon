package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/redis/go-redis/v9"
)

var defaultTimeout = 5 * time.Second

func SetTimeout(timeout time.Duration) {
	defaultTimeout = timeout
}

// NewRedisCache 创建redis缓存
func NewRedisCache(client *redis.Client) agent.Cache {
	return &redisCache{client: client}
}

type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func (r *redisCache) WithContext(ctx context.Context) agent.Cache {
	r.ctx = context.WithoutCancel(ctx)
	return r
}

func (r *redisCache) SetNX(key string, value any, expiration time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return r.client.SetNX(ctx, key, value, expiration).Val()
}

func (r *redisCache) Exists(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return r.client.Exists(ctx, key).Val() > 0
}

// Get 获取缓存
//
// 需要value实现 encoding.BinaryUnmarshaler
func (r *redisCache) Get(key string, value any) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	if err := r.client.Get(ctx, key).Scan(value); err != nil {
		if errors.Is(err, redis.Nil) {
			return agent.NoCache
		}
		return err
	}
	return nil
}

func (r *redisCache) Set(key string, value any, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, bytes, expiration).Err()
}

func (r *redisCache) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	return r.client.Del(ctx, key).Err()
}
