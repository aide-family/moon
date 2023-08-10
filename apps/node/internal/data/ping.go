package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/apps/node/internal/conf"
)

type (
	PingRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

func (l *PingRepo) Check(ctx context.Context) (*conf.Env, error) {
	_, span := l.tr.Start(ctx, "Check")
	defer span.End()

	return conf.Get().GetEnv(), nil
}

var _ biz.IPingRepo = (*PingRepo)(nil)

func NewPingRepo(data *Data, logger log.Logger) *PingRepo {
	return &PingRepo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", "data/Ping")),
		tr:     otel.Tracer("data/Ping"),
	}
}
