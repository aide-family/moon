package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/demo/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/demo/internal/data"
	"github.com/aide-family/moon/pkg/util/conn"
)

// NewCacheRepo new cache repo.
func NewCacheRepo(data *data.Data) repository.CacheRepo {
	return &cacheRepoImpl{data: data}
}

type cacheRepoImpl struct {
	data *data.Data
}

func (l *cacheRepoImpl) Cacher() conn.Cache {
	return l.data.GetCacher()
}
