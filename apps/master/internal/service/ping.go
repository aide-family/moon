package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api"
)

type (
	IPingLogic interface {
		Check(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error)
	}

	PingService struct {
		pb.UnimplementedPingServer

		logger *log.Helper
		logic  IPingLogic
	}
)

var _ pb.PingServer = (*PingService)(nil)

func NewPingService(logic IPingLogic, logger log.Logger) *PingService {
	return &PingService{logic: logic, logger: log.NewHelper(log.With(logger, "module", pingModuleName))}
}

func (l *PingService) Check(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	ctx, span := otel.Tracer(pingModuleName).Start(ctx, "PingService.Check")
	defer span.End()
	return l.logic.Check(ctx, req)
}
