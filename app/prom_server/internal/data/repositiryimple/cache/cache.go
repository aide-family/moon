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
	repository.UnimplementedCacheRepo

	log *log.Helper

	data *data.Data
}

func NewCacheRepo(data *data.Data, logger log.Logger) repository.CacheRepo {
	return &cacheRepoImpl{
		log:  log.NewHelper(log.With(logger, "module", "data.cache")),
		data: data,
	}
}

func (l *cacheRepoImpl) Client() (*redis.Client, error) {
	return l.data.Client(), nil
}

func (l *cacheRepoImpl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	client, err := l.Client()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration).Err()
}
