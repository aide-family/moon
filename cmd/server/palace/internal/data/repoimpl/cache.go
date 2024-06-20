package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/conn"
)

func NewCacheRepository(data *data.Data) repository.Cache {
	return &cacheRepositoryImpl{data: data}
}

type cacheRepositoryImpl struct {
	data *data.Data
}

func (l *cacheRepositoryImpl) Cacher() conn.Cache {
	return l.data.GetCacher()
}
