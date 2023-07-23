package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/node/internal/biz"
)

type (
	PullRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *PullRepo) V1(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

var _ biz.IPullRepo = (*PullRepo)(nil)

func NewPullRepo(data *Data, logger log.Logger) *PullRepo {
	return &PullRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Pull"))}
}
