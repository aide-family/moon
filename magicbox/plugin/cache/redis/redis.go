// Package redis is a cache driver for redis.
package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aide-family/magicbox/plugin/cache"
)

var _ cache.Driver = (*initializer)(nil)
var _ cache.Interface = (*redisCache)(nil)

// CacheDriver returns a new cache driver.
func CacheDriver(cli *redis.Client) cache.Driver {
	return &initializer{cli: cli}
}

type initializer struct {
	cli *redis.Client
}

// New implements cache.Driver.
func (i *initializer) New(ctx context.Context) (cache.Interface, error) {
	return &redisCache{
		cli: i.cli,
	}, nil
}

type redisCache struct {
	cli *redis.Client
}

// Close implements cache.Interface.
func (r *redisCache) Close() error {
	return r.cli.Close()
}

// Del implements cache.Interface.
func (r *redisCache) Del(ctx context.Context, key cache.K) error {
	return r.cli.Del(ctx, key.String()).Err()
}

// Exists implements cache.Interface.
func (r *redisCache) Exists(ctx context.Context, key cache.K) (bool, error) {
	res, err := r.cli.Exists(ctx, key.String()).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

// Get implements cache.Interface.
func (r *redisCache) Get(ctx context.Context, key cache.K) (string, error) {
	return r.cli.Get(ctx, key.String()).Result()
}

// HDel implements cache.Interface.
func (r *redisCache) HDel(ctx context.Context, key cache.K, field string) error {
	return r.cli.HDel(ctx, key.String(), field).Err()
}

// HExists implements cache.Interface.
func (r *redisCache) HExists(ctx context.Context, key cache.K, field string) (bool, error) {
	res, err := r.cli.HExists(ctx, key.String(), field).Result()
	if err != nil {
		return false, err
	}
	return res, nil
}

// HGet implements cache.Interface.
func (r *redisCache) HGet(ctx context.Context, key cache.K, field string) (string, error) {
	return r.cli.HGet(ctx, key.String(), field).Result()
}

// HMGet implements cache.Interface.
func (r *redisCache) HMGet(ctx context.Context, key cache.K, fields ...string) ([][]byte, error) {
	res, err := r.cli.HMGet(ctx, key.String(), fields...).Result()
	if err != nil {
		return nil, err
	}
	resStr := make([][]byte, 0, len(res))
	for _, v := range res {
		switch val := v.(type) {
		case string:
			resStr = append(resStr, []byte(val))
		case []byte:
			resStr = append(resStr, val)
		case nil:
			resStr = append(resStr, []byte{})
		default:
			return nil, fmt.Errorf("invalid type: %T", val)
		}
	}
	return resStr, nil
}

// HMSet implements cache.Interface.
func (r *redisCache) HMSet(ctx context.Context, key cache.K, fields map[string]string) error {
	return r.cli.HMSet(ctx, key.String(), fields).Err()
}

// HSet implements cache.Interface.
func (r *redisCache) HSet(ctx context.Context, key cache.K, field string, value string) error {
	return r.cli.HSet(ctx, key.String(), field, value).Err()
}

// IncMax implements cache.Interface.
func (r *redisCache) IncMax(ctx context.Context, key cache.K, max int, ttl time.Duration) (bool, error) {
	res, err := r.cli.Eval(ctx, `
		local key = KEYS[1]
		local max = tonumber(ARGV[1])
		local expire = tonumber(ARGV[2])
		local current = tonumber(redis.call("get", key))
		if current == nil then
			redis.call("set", key, 1)
			redis.call("expire", key, expire)
			return 1
		end
		if current >= max then
			return current
		end
		redis.call("incr", key)
		return current
	`, []string{key.String()}, max, int(ttl.Seconds())).Int()
	if err != nil {
		return false, err
	}
	return res < max, nil
}

// Lock implements cache.Interface.
func (r *redisCache) Lock(ctx context.Context, key cache.K, ttl time.Duration) (bool, error) {
	return r.cli.SetNX(ctx, key.String(), 1, ttl).Result()
}

// Set implements cache.Interface.
func (r *redisCache) Set(ctx context.Context, key cache.K, value string, ttl time.Duration) error {
	return r.cli.Set(ctx, key.String(), value, ttl).Err()
}

// Unlock implements cache.Interface.
func (r *redisCache) Unlock(ctx context.Context, key cache.K) error {
	return r.cli.Del(ctx, key.String()).Err()
}

// ZAdd implements cache.Interface.
func (r *redisCache) ZAdd(ctx context.Context, key cache.K, score float64, member string) error {
	return r.cli.ZAdd(ctx, key.String(), redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRange implements cache.Interface.
func (r *redisCache) ZRange(ctx context.Context, key cache.K, start int, stop int) ([]string, error) {
	return r.cli.ZRange(ctx, key.String(), int64(start), int64(stop)).Result()
}

// ZRangeByScore implements cache.Interface.
func (r *redisCache) ZRangeByScore(ctx context.Context, key cache.K, min float64, max float64) ([]string, error) {
	return r.cli.ZRangeByScore(ctx, key.String(), &redis.ZRangeBy{
		Min:    strconv.FormatFloat(min, 'f', -1, 64),
		Max:    strconv.FormatFloat(max, 'f', -1, 64),
		Offset: 0,
		Count:  0,
	}).Result()
}

// ZRem implements cache.Interface.
func (r *redisCache) ZRem(ctx context.Context, key cache.K, member string) error {
	return r.cli.ZRem(ctx, key.String(), member).Err()
}

// ZRemRangeByScore implements cache.Interface.
func (r *redisCache) ZRemRangeByScore(ctx context.Context, key cache.K, min float64, max float64) error {
	return r.cli.ZRemRangeByScore(ctx, key.String(), strconv.FormatFloat(min, 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64)).Err()
}
