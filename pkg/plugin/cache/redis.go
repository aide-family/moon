package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/pkg/config"
)

var _ Cache = (*redisCache)(nil)

// NewRedisCache create a redis cache.
func NewRedisCache(cli *redis.Client, driver config.Cache_Driver) Cache {
	return &redisCache{client: cli, driver: driver}
}

type redisCache struct {
	client *redis.Client
	driver config.Cache_Driver
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

func (r *redisCache) Client() *redis.Client {
	return r.client
}

func (r *redisCache) Driver() config.Cache_Driver {
	return r.driver
}

func (r *redisCache) IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error) {
	return r.client.Eval(ctx, `
local key = KEYS[1]
local max = tonumber(ARGV[1])
local expire = tonumber(ARGV[2])
local current = tonumber(redis.call('get', key))
if current == nil then
	redis.call('set', key, 1)
	redis.call('expire', key, expire)
	return 1
end
if current < max then
	redis.call('incr', key)
	return 1
end
return 0
`, []string{key}, max, expiration/time.Second).Bool()
}

func (r *redisCache) DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error) {
	return r.client.Eval(ctx, `
local key = KEYS[1]
local min = tonumber(ARGV[1])
local expire = tonumber(ARGV[2])
local current = tonumber(redis.call('get', key))
if current == nil then
	redis.call('set', key, 1)
	redis.call('expire', key, expire)
	return 1
end
if current > min then
	redis.call('decr', key)
	return 1
end
return 0
`, []string{key}, min, expiration/time.Second).Bool()
}
