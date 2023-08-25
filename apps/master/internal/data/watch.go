package data

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/pkg/alert"
)

type (
	WatchRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IWatchRepo = (*WatchRepo)(nil)

func NewWatchRepo(data *Data, logger log.Logger) *WatchRepo {
	return &WatchRepo{data: data, logger: log.NewHelper(log.With(logger, "module", watchModuleName))}
}

func (l *WatchRepo) V1(_ context.Context) string {
	return "WatchRepo.V1"
}

func (l *WatchRepo) WatchAlert(ctx context.Context, req *alert.Data) error {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "WatchRepo.WatchAlert")
	defer span.End()
	marshal, err := json.Marshal(req)
	if err != nil {
		return err
	}
	// TODO 落库
	fmt.Println(string(marshal))
	return nil
}
