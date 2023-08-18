package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/strategy/v1/pull"
)

type (
	IPullLogic interface {
		Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error)
		Datasources(ctx context.Context, req *pb.DatasourcesRequest) (*pb.DatasourcesReply, error)
	}

	PullService struct {
		pb.UnimplementedPullServer

		logger *log.Helper
		logic  IPullLogic
	}
)

var _ pb.PullServer = (*PullService)(nil)

func NewPullService(logic IPullLogic, logger log.Logger) *PullService {
	return &PullService{logic: logic, logger: log.NewHelper(log.With(logger, loadModuleName))}
}

func (l *PullService) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := otel.Tracer(pullModuleName).Start(ctx, "PullService.Strategies")
	defer span.End()
	return l.logic.Strategies(ctx, req)
}

func (l *PullService) Datasources(ctx context.Context, req *pb.DatasourcesRequest) (*pb.DatasourcesReply, error) {
	ctx, span := otel.Tracer(pullModuleName).Start(ctx, "PullService.Datasources")
	defer span.End()
	return l.logic.Datasources(ctx, req)
}
