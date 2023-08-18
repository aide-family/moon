package data

import (
	"context"

	"go.opentelemetry.io/otel"
	"prometheus-manager/apps/node/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/node/internal/conf"
)

type (
	PingRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IPingRepo = (*PingRepo)(nil)

func NewPingRepo(data *Data, logger log.Logger) *PingRepo {
	return &PingRepo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", pingModuleName)),
	}
}

func (l *PingRepo) Check(ctx context.Context) (*conf.Env, error) {
	_, span := otel.Tracer(pingModuleName).Start(ctx, "PingRepo.Check")
	defer span.End()

	return conf.Get().GetEnv(), nil
}
