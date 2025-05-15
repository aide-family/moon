package impl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/rabbit/internal/data"
)

func NewCacheRepo(d *data.Data, logger log.Logger) repository.Cache {
	return &cacheImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.cache")),
	}
}

type cacheImpl struct {
	*data.Data

	helper *log.Helper
}

func (c *cacheImpl) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.Data.GetCache().Client().SetNX(ctx, key, 1, expiration).Result()
}

func (c *cacheImpl) Unlock(ctx context.Context, key string) error {
	return c.Data.GetCache().Client().Del(ctx, key).Err()
}
