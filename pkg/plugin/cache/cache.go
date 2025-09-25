// Package cache is a cache package for kratos.
package cache

import (
	"context"
	"encoding"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/pkg/config"
)

type Object interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	UniqueKey() string
}

// Cache is the interface of cache.
type Cache interface {
	// Close the cache client.
	Close() error
	// Client Get the redis client.
	Client() *redis.Client
	// Driver Get the cache driver.
	Driver() config.Cache_Driver
	// IncMax Increase the cached value, and if it exceeds max, return false.
	IncMax(ctx context.Context, key string, max int64, expiration time.Duration) (bool, error)
	// DecMin Decrease the cached value, and if it is less than min, return false.
	DecMin(ctx context.Context, key string, min int64, expiration time.Duration) (bool, error)
}

// NewCache create a cache client.
func NewCache(c *config.Cache) (Cache, error) {
	var (
		cli *redis.Client
		err error
	)
	switch c.GetDriver() {
	case config.Cache_REDIS:
		cli = newRedisWithConfig(c)
	default:
		cli, err = newRedisWithMiniRedis(c)
		if err != nil {
			return nil, err
		}
	}
	if err = cli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return NewRedisCache(cli, c.GetDriver()), nil
}

func newRedisWithConfig(c *config.Cache) *redis.Client {
	return redis.NewClient(newDefaultOptions(c))
}

func newRedisWithMiniRedis(c *config.Cache) (*redis.Client, error) {
	cli, err := miniredis.Run()
	if err != nil {
		return nil, err
	}
	options := newDefaultOptions(c)
	options.Network = "tcp"
	options.Addr = cli.Addr()
	options.Username = ""
	options.Password = ""
	return redis.NewClient(options), nil
}

func newDefaultOptions(c *config.Cache) *redis.Options {
	return &redis.Options{
		Network:      c.GetNetwork(),
		Addr:         c.GetAddr(),
		Username:     c.GetUsername(),
		Password:     c.GetPassword(),
		DB:           int(c.GetDb()),
		DialTimeout:  c.GetDialTimeout().AsDuration(),
		ReadTimeout:  c.GetReadTimeout().AsDuration(),
		WriteTimeout: c.GetWriteTimeout().AsDuration(),
	}
}
