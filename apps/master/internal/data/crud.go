package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/dal/model"
)

type (
	CrudRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *CrudRepo) CreateStrategies(ctx context.Context, m *model.PromNodeDirFileGroupStrategy) error {
	//TODO implement me
	panic("implement me")
}

var _ biz.ICrudRepo = (*CrudRepo)(nil)

func NewCrudRepo(data *Data, logger log.Logger) *CrudRepo {
	return &CrudRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Crud"))}
}
