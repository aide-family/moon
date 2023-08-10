package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"prometheus-manager/apps/node/internal/biz"
)

type (
	PullRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

func (l *PullRepo) PullStrategies(ctx context.Context) (*biz.StrategyLoad, error) {
	_, span := l.tr.Start(ctx, "PullStrategies")
	defer span.End()

	return &biz.StrategyLoad{
		StrategyDirs: strategies,
		LoadTime:     loadTime,
	}, nil
}

func (l *PullRepo) V1(ctx context.Context) string {
	ctx, span := l.tr.Start(ctx, "showVersion")
	defer span.End()
	return "version is v1"
}

var _ biz.IPullRepo = (*PullRepo)(nil)

func NewPullRepo(data *Data, logger log.Logger) *PullRepo {
	return &PullRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Pull")), tr: otel.Tracer("data/Pull")}
}
