package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/biz"
)

type (
	PingRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *PingRepo) V1(ctx context.Context) string {
	//TODO implement me
	panic("implement me")
}

var _ biz.IPingRepo = (*PingRepo)(nil)

func NewPingRepo(data *Data, logger log.Logger) *PingRepo {
	return &PingRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Ping"))}
}
