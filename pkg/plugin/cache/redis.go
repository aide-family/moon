package cache

import (
	"context"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

var _ ICacher = (*redisCacher)(nil)

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

func (r *redisCacher) Keys(ctx context.Context, prefix string) ([]string, error) {
	return r.client.Keys(ctx, prefix).Result()
}

func (r *redisCacher) DelKeys(ctx context.Context, prefix string) error {
	keys, err := r.Keys(ctx, prefix)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisCacher) GetInt64(ctx context.Context, key string) (int64, error) {
	return r.client.Get(ctx, key).Int64()
}

func (r *redisCacher) SetInt64(ctx context.Context, key string, value int64, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCacher) Close() error {
	if r == nil {
		return nil
	}
	return r.client.Close()
}

func (r *redisCacher) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacher) Exist(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *redisCacher) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCacher) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCacher) Inc(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisCacher) Dec(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return r.client.Decr(ctx, key).Result()
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

func (r *redisCacher) GetFloat64(ctx context.Context, key string) (float64, error) {
	return r.client.Get(ctx, key).Float64()
}

func (r *redisCacher) SetFloat64(ctx context.Context, key string, value float64, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCacher) GetObject(ctx context.Context, key string, obj IObjectSchema) error {
	return r.client.Get(ctx, key).Scan(obj)
}

func (r *redisCacher) SetObject(ctx context.Context, key string, obj IObjectSchema, expiration time.Duration) error {
	return r.client.Set(ctx, key, obj, expiration).Err()
}

func (r *redisCacher) GetBool(ctx context.Context, key string) (bool, error) {
	return r.client.Get(ctx, key).Bool()
}

func (r *redisCacher) SetBool(ctx context.Context, key string, value bool, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *redisCacher) SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}
