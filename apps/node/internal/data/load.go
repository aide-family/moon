package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"prometheus-manager/api/strategy"
	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/pkg/util/strategyload"
	"time"
)

type (
	LoadRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

func (l *LoadRepo) LoadStrategy(ctx context.Context, path []string) error {
	_, span := l.tr.Start(ctx, "LoadStrategy")
	defer span.End()

	var tmpStrategies []*strategy.StrategyDir
	for _, p := range path {
		c := strategyload.NewStrategy(file.NewSource(p))
		var strategyList []*strategy.Strategy
		if err := c.Scan(&strategyList); err != nil {
			l.logger.Errorf("c.Scan err: %v", err)
		}

		tmpStrategies = append(tmpStrategies, &strategy.StrategyDir{
			Dir:        p,
			Strategies: strategyList,
		})
	}

	strategies = tmpStrategies
	loadTime = time.Now()

	return nil
}

func (l *LoadRepo) V1(ctx context.Context) (string, error) {
	ctx, span := l.tr.Start(ctx, "showVersion")
	defer span.End()
	return "version is v1", nil
}

var _ biz.ILoadRepo = (*LoadRepo)(nil)

func NewLoadRepo(data *Data, logger log.Logger) *LoadRepo {
	return &LoadRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Load")), tr: otel.Tracer("data/Load")}
}
