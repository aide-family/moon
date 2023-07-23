package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/node/internal/biz"
)

type (
	PushRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *PushRepo) V1(ctx context.Context) (string, error) {
	//TODO implement me
	panic("implement me")
}

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Push"))}
}
