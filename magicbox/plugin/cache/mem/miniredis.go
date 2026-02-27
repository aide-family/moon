// Package mem is a cache driver for memory.
package mem

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/aide-family/magicbox/plugin/cache"
)

var _ cache.Driver = (*initializer)(nil)
var _ cache.Interface = (*miniRedisCache)(nil)

// CacheDriver returns a new cache driver.
func CacheDriver() cache.Driver {
	return &initializer{}
}

type initializer struct{}

// New implements cache.Driver.
func (i *initializer) New(ctx context.Context) (cache.Interface, error) {
	cli, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	return &miniRedisCache{
		cli: redis.NewClient(&redis.Options{
			Addr: cli.Addr(),
		}),
	}, nil
}

type miniRedisCache struct {
	cli *redis.Client
}

// Close implements cache.Interface.
func (m *miniRedisCache) Close() error {
	return m.cli.Close()
}

// Del implements cache.Interface.
func (m *miniRedisCache) Del(ctx context.Context, key cache.K) error {
	return m.cli.Del(ctx, key.String()).Err()
}

// Exists implements cache.Interface.
func (m *miniRedisCache) Exists(ctx context.Context, key cache.K) (bool, error) {
	res, err := m.cli.Exists(ctx, key.String()).Result()
	if err != nil {
		return false, err
	}
	return res > 0, nil
}

// Get implements cache.Interface.
func (m *miniRedisCache) Get(ctx context.Context, key cache.K) (string, error) {
	return m.cli.Get(ctx, key.String()).Result()
}

// Set implements cache.Interface.
func (m *miniRedisCache) Set(ctx context.Context, key cache.K, value string, ttl time.Duration) error {
	return m.cli.Set(ctx, key.String(), value, ttl).Err()
}

// HDel implements cache.Interface.
func (m *miniRedisCache) HDel(ctx context.Context, key cache.K, field string) error {
	return m.cli.HDel(ctx, key.String(), field).Err()
}

// HExists implements cache.Interface.
func (m *miniRedisCache) HExists(ctx context.Context, key cache.K, field string) (bool, error) {
	res, err := m.cli.HExists(ctx, key.String(), field).Result()
	if err != nil {
		return false, err
	}
	return res, nil
}

// HGet implements cache.Interface.
func (m *miniRedisCache) HGet(ctx context.Context, key cache.K, field string) (string, error) {
	return m.cli.HGet(ctx, key.String(), field).Result()
}

// HMGet implements cache.Interface.
func (m *miniRedisCache) HMGet(ctx context.Context, key cache.K, fields ...string) ([][]byte, error) {
	res, err := m.cli.HMGet(ctx, key.String(), fields...).Result()
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
func (m *miniRedisCache) HMSet(ctx context.Context, key cache.K, fields map[string]string) error {
	return m.cli.HMSet(ctx, key.String(), fields).Err()
}

// HSet implements cache.Interface.
func (m *miniRedisCache) HSet(ctx context.Context, key cache.K, field string, value string) error {
	return m.cli.HSet(ctx, key.String(), field, value).Err()
}

// IncMax implements cache.Interface.
func (m *miniRedisCache) IncMax(ctx context.Context, key cache.K, max int, ttl time.Duration) (bool, error) {
	res, err := m.cli.Eval(ctx, `
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
func (m *miniRedisCache) Lock(ctx context.Context, key cache.K, ttl time.Duration) (bool, error) {
	return m.cli.SetNX(ctx, key.String(), 1, ttl).Result()
}

// Unlock implements cache.Interface.
func (m *miniRedisCache) Unlock(ctx context.Context, key cache.K) error {
	return m.cli.Del(ctx, key.String()).Err()
}

// ZAdd implements cache.Interface.
func (m *miniRedisCache) ZAdd(ctx context.Context, key cache.K, score float64, member string) error {
	return m.cli.ZAdd(ctx, key.String(), redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRange implements cache.Interface.
func (m *miniRedisCache) ZRange(ctx context.Context, key cache.K, start int, stop int) ([]string, error) {
	return m.cli.ZRange(ctx, key.String(), int64(start), int64(stop)).Result()
}

// ZRangeByScore implements cache.Interface.
func (m *miniRedisCache) ZRangeByScore(ctx context.Context, key cache.K, min float64, max float64) ([]string, error) {
	return m.cli.ZRangeByScore(ctx, key.String(), &redis.ZRangeBy{
		Min:    strconv.FormatFloat(min, 'f', -1, 64),
		Max:    strconv.FormatFloat(max, 'f', -1, 64),
		Offset: 0,
		Count:  0,
	}).Result()
}

// ZRem implements cache.Interface.
func (m *miniRedisCache) ZRem(ctx context.Context, key cache.K, member string) error {
	return m.cli.ZRem(ctx, key.String(), member).Err()
}

// ZRemRangeByScore implements cache.Interface.
func (m *miniRedisCache) ZRemRangeByScore(ctx context.Context, key cache.K, min float64, max float64) error {
	return m.cli.ZRemRangeByScore(ctx, key.String(), strconv.FormatFloat(min, 'f', -1, 64), strconv.FormatFloat(max, 'f', -1, 64)).Err()
}
