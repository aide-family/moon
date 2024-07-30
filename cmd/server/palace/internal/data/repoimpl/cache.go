package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/conn"
)

// NewCacheRepository 创建缓存操作
func NewCacheRepository(data *data.Data) repository.Cache {
	return &cacheRepositoryImpl{data: data}
}

type cacheRepositoryImpl struct {
	data *data.Data
}

// Cacher 获取缓存实例
func (l *cacheRepositoryImpl) Cacher() conn.Cache {
	return l.data.GetCacher()
}
