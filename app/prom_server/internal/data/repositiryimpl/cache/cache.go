package cache

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/cache"
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

func (l *cacheRepoImpl) Client() (cache.GlobalCache, error) {
	return l.data.Cache(), nil
}

func (l *cacheRepoImpl) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	client, err := l.Client()
	if err != nil {
		return err
	}
	return client.Set(ctx, key, value, expiration)
}
