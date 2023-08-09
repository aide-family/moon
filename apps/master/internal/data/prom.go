package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
)

type (
	PromRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IPromRepo = (*PromRepo)(nil)

func NewPromRepo(data *Data, logger log.Logger) *PromRepo {
	return &PromRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Prom"))}
}

func (p PromRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}
