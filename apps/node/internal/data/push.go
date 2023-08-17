package data

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"prometheus-manager/pkg/util/dir"

	"prometheus-manager/api/strategy"

	"prometheus-manager/pkg/util/strategystore"

	"prometheus-manager/apps/node/internal/biz"
)

type (
	PushRepo struct {
		logger *log.Helper
		data   *Data
		tr     trace.Tracer
	}
)

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{
		data:   data,
		logger: log.NewHelper(log.With(logger, "module", "data/Push")),
		tr:     otel.Tracer("data/Push"),
	}
}

func (l *PushRepo) StoreStrategy(ctx context.Context, strategyDirList []*strategy.StrategyDir) (*biz.StoreStrategyResult, error) {
	_, span := l.tr.Start(ctx, "StoreStrategy")
	defer span.End()

	store := strategystore.NewStrategy(strategyDirList, l.logger)

	storeStrategy, err := store.StoreStrategy()
	if err != nil {
		return nil, err
	}

	return &biz.StoreStrategyResult{
		SuccessCount: int64(len(strategyDirList)) - int64(len(storeStrategy)),
		FailedCount:  int64(len(storeStrategy)),
		StrategyDirs: storeStrategy,
	}, nil
}

func (l *PushRepo) RemoveStrategy(ctx context.Context, files []string) error {
	_, span := l.tr.Start(ctx, "PushRepo.RemoveStrategy")
	defer span.End()
	return dir.RemoveAllYamlFilename(files...)
}

func (l *PushRepo) V1(ctx context.Context) string {
	ctx, span := l.tr.Start(ctx, "showVersion")
	defer span.End()
	return "version is v1"
}
