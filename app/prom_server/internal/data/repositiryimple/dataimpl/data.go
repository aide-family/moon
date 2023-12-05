package dataimpl

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"prometheus-manager/api/perrors"

	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
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

func (l *dataRepoImpl) Client() (*redis.Client, error) {
	client := l.d.Client()
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
