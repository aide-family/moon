package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/alert/v1"
)

type (
	IWatchLogic interface {
		WatchAlert(ctx context.Context, req *pb.WatchRequest) (*pb.WatchReply, error)
	}

	WatchService struct {
		pb.UnimplementedWatchServer

		logger *log.Helper
		logic  IWatchLogic
	}
)

var _ pb.WatchServer = (*WatchService)(nil)

func NewWatchService(logic IWatchLogic, logger log.Logger) *WatchService {
	return &WatchService{logic: logic, logger: log.NewHelper(log.With(logger, "module", alarmPageModuleName))}
}

func (l *WatchService) WatchAlert(ctx context.Context, req *pb.WatchRequest) (*pb.WatchReply, error) {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "WatchService.WatchAlert")
	defer span.End()
	return l.logic.WatchAlert(ctx, req)
}
