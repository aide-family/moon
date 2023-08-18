package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api/strategy"

	"prometheus-manager/pkg/util/strategyload"

	"prometheus-manager/apps/node/internal/biz"
)

type (
	LoadRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.ILoadRepo = (*LoadRepo)(nil)

func NewLoadRepo(data *Data, logger log.Logger) *LoadRepo {
	return &LoadRepo{data: data, logger: log.NewHelper(log.With(logger, "module", loadModuleName))}
}

func (l *LoadRepo) LoadStrategy(ctx context.Context, path []string) error {
	ctx, span := otel.Tracer(loadModuleName).Start(ctx, "LoadStrategy")
	defer span.End()

	var tmpStrategies []*strategy.StrategyDir
	for _, p := range path {
		c := strategyload.NewStrategy(file.NewSource(p))
		var strategyList []*strategy.Strategy
		if err := c.Scan(ctx, &strategyList); err != nil {
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

func (l *LoadRepo) V1(ctx context.Context) string {
	_, span := otel.Tracer(loadModuleName).Start(ctx, "LoadRepo.V1")
	defer span.End()
	return "version is v1"
}
