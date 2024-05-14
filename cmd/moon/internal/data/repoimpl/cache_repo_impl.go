package repoimpl

import (
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/repo"
	"github.com/aide-cloud/moon/cmd/moon/internal/data"
	"github.com/aide-cloud/moon/pkg/conn"
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
