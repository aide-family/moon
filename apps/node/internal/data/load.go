package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/node/internal/biz"
)

type (
	LoadRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *LoadRepo) V1(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

var _ biz.ILoadRepo = (*LoadRepo)(nil)

func NewLoadRepo(data *Data, logger log.Logger) *LoadRepo {
	return &LoadRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Load"))}
}
