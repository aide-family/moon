package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/apps/master/internal/conf"
)

type (
	PingRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IPingRepo = (*PingRepo)(nil)

func NewPingRepo(data *Data, logger log.Logger) *PingRepo {
	return &PingRepo{data: data, logger: log.NewHelper(log.With(logger, "module", pingModuleName))}
}

func (l *PingRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer(pingModuleName).Start(ctx, "PingRepo.V1")
	defer span.End()
	l.logger.WithContext(ctx).Infof("PingRepo.V1")
	return conf.Get().GetEnv().GetVersion()
}
