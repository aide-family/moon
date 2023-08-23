package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/apps/node/internal/biz"
)

type (
	AlertRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IAlertRepo = (*AlertRepo)(nil)

func NewAlertRepo(data *Data, logger log.Logger) *AlertRepo {
	return &AlertRepo{data: data, logger: log.NewHelper(log.With(logger, "module", alertModuleName))}
}

func (l *AlertRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer(alertModuleName).Start(ctx, "AlertRepo.V1")
	defer span.End()
	return "AlertRepo.V1"
}
