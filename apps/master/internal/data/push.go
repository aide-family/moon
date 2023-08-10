package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/apps/master/internal/biz"
)

type (
	PushRepo struct {
		logger *log.Helper
		data   *Data
	}
)

func (l *PushRepo) V1(ctx context.Context) string {
	_, span := otel.Tracer("data").Start(ctx, "PushRepo.V1")
	defer span.End()
	return "PushRepo.V1"
}

var _ biz.IPushRepo = (*PushRepo)(nil)

func NewPushRepo(data *Data, logger log.Logger) *PushRepo {
	return &PushRepo{data: data, logger: log.NewHelper(log.With(logger, "module", "data/Push"))}
}
