package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"prometheus-manager/apps/node/internal/biz"
)

type (
	LoadRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

func (l *LoadRepo) V1(ctx context.Context) (string, error) {
	ctx, span := l.tr.Start(ctx, "showVersion")
	defer span.End()
	return "version is v1", nil
}

var _ biz.ILoadRepo = (*LoadRepo)(nil)

func NewLoadRepo(data *Data, logger log.Logger) *LoadRepo {
	return &LoadRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Load")), tr: otel.Tracer("data/Load")}
}
