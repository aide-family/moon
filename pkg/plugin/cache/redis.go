package cache

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// NewRedisCacher creates a new redis cacher
func NewRedisCacher(cli *redis.Client) ICacher {
	return &redisCacher{client: cli}
}

// NewRedisCacherByMiniRedis creates a new redis cacher by mini redis
func NewRedisCacherByMiniRedis(cli *miniredis.Miniredis) ICacher {
	c := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    cli.Addr(),
	})
	return &redisCacher{client: c}
}

type (
	redisCacher struct {
		client *redis.Client
	}
)

// Client implements ICacher.
func (r *redisCacher) Client() *redis.Client {
	return r.client
}

func (r *redisCacher) Close() error {
	if r == nil {
		return nil
	}
	return r.client.Close()
}

func (r *redisCacher) IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error) {
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

func (r *redisCacher) DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error) {
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
