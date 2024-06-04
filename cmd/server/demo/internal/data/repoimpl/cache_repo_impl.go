package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/demo/internal/biz/repo"
	"github.com/aide-family/moon/cmd/server/demo/internal/data"
	"github.com/aide-family/moon/pkg/conn"
)

func NewCacheRepo(data *data.Data) repo.CacheRepo {
	return &cacheRepoImpl{data: data}
}

type cacheRepoImpl struct {
	data *data.Data
}

func (l *cacheRepoImpl) Cacher() conn.Cache {
	return l.data.GetCacher()
}
