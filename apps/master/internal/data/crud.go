package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
)

type (
	CrudRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *CrudRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ biz.ICrudRepo = (*CrudRepo)(nil)

func NewCrudRepo(data *Data, logger log.Logger) *CrudRepo {
	return &CrudRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Crud"))}
}
