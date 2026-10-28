package data

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/aide-family/moon/app/merak/internal/conf"
	"github.com/aide-family/moon/app/merak/internal/data/repository"
	"github.com/aide-family/moon/pkg/lock"
	"github.com/aide-family/moon/pkg/merr"
)

type lockClient struct {
	rdb   *redis.Client
	locks *lock.Client
}

func NewLocker(c *conf.Redis, logger *slog.Logger) (repository.Locker, func(), error) {
	if c == nil || c.GetAddr() == "" {
		return nil, nil, merr.Server("redis addr is empty")
	}
	rdb := lock.NewRedisClient(c.GetAddr(), c.GetPassword())
	cleanup := func() {
		if err := rdb.Close(); err != nil {
			logger.Error(fmt.Sprintf("close redis: %v", err))
		}
	}
	return &lockClient{rdb: rdb, locks: lock.NewClient(rdb, "merak")}, cleanup, nil
}

func (c *lockClient) Available(ctx context.Context) error {
	return c.locks.Available(ctx)
}

func (c *lockClient) Acquire(ctx context.Context, key string, ttl time.Duration) (repository.Hold, error) {
	hold, err := c.locks.Acquire(ctx, key, ttl)
	if err != nil {
		return nil, err
	}
	return hold, nil
}
