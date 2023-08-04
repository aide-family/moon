package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/node"
)

type (
	IPushLogic interface {
		Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error)
	}

	PushService struct {
		pb.UnimplementedPushServer

		logger *log.Helper
		logic  IPushLogic
	}
)

var _ pb.PushServer = (*PushService)(nil)

func NewPushService(logic IPushLogic, logger log.Logger) *PushService {
	return &PushService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Push"))}
}

func (l *PushService) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PushService.Call")
	defer span.End()
	return l.logic.Call(ctx, req)
}
