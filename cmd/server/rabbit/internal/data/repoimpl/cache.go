package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/data"
	"github.com/aide-family/moon/pkg/plugin/cache"
)

// NewCacheRepo 创建缓存操作
func NewCacheRepo(data *data.Data) repository.CacheRepo {
	return &cacheRepoImpl{data: data}
}

type cacheRepoImpl struct {
	data *data.Data
}

func (l *cacheRepoImpl) Cacher() cache.ICacher {
	return l.data.GetCacher()
}
