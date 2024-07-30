package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/util/conn"
)

// NewCacheRepo 实例化缓存仓库
func NewCacheRepo(data *data.Data) repository.CacheRepo {
	return &cacheRepoImpl{data: data}
}

type cacheRepoImpl struct {
	data *data.Data
}

// Cacher 获取缓存仓库
func (l *cacheRepoImpl) Cacher() conn.Cache {
	return l.data.GetCacher()
}
