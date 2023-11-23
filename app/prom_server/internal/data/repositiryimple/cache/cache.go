package cache

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.CacheRepo = (*cacheRepoImpl)(nil)

type cacheRepoImpl struct {
	log *log.Helper

	data *data.Data
}

func NewCacheRepo(data *data.Data, logger log.Logger) repository.CacheRepo {
	return &cacheRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.cache")),
		data: data,
	}
}

func (l *cacheRepoImpl) Client() *redis.Client {
	return l.data.Client()
}

func (l *cacheRepoImpl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return l.Client().Set(ctx, key, value, expiration).Err()
}
