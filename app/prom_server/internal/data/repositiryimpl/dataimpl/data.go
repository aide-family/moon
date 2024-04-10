package dataimpl

import (
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"github.com/aide-family/moon/pkg/util/cache"

	"github.com/aide-family/moon/api/perrors"

	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
)

var _ repository.DataRepo = (*dataRepoImpl)(nil)

type dataRepoImpl struct {
	repository.UnimplementedDataRepo

	log *log.Helper
	d   *data.Data
}

func (l *dataRepoImpl) DB() (*gorm.DB, error) {
	db := l.d.DB()
	if db == nil {
		return nil, perrors.ErrorUnknown("db is nil")
	}
	return db, nil
}

func (l *dataRepoImpl) Cache() (cache.GlobalCache, error) {
	client := l.d.Cache()
	if client == nil {
		return nil, perrors.ErrorUnknown("client is nil")
	}
	return client, nil
}

func NewDataRepo(data *data.Data, logger log.Logger) repository.DataRepo {
	return &dataRepoImpl{
		log: log.NewHelper(log.With(logger, "module", "data")),
		d:   data,
	}
}
