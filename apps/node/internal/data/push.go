package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"prometheus-manager/apps/node/internal/biz"
)

type (
	PushRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

func (l *PushRepo) V1(ctx context.Context) (string, error) {
	ctx, span := l.tr.Start(ctx, "showVersion")
	defer span.End()
	return "version is v1", nil
}

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Push")), tr: otel.Tracer("data/Push")}
}
