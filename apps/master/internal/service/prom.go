package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/prom/v1"
)

type (
	IPromV1Logic interface {
		CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error)
		UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error)
		UpdateGroupsStatus(ctx context.Context, req *pb.UpdateGroupsStatusRequest) (*pb.UpdateGroupsStatusReply, error)
		DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error)
		GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error)
		ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error)
		CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error)
		UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error)
		UpdateStrategiesStatus(ctx context.Context, req *pb.UpdateStrategiesStatusRequest) (*pb.UpdateStrategiesStatusReply, error)
		DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error)
		GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error)
		ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error)
	}

	PromV1Service struct {
		pb.UnimplementedPromServer

		logger *log.Helper
		logic  IPromV1Logic
	}
)

var _ pb.PromServer = (*PromV1Service)(nil)

func NewPromService(logic IPromV1Logic, logger log.Logger) *PromV1Service {
	return &PromV1Service{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Prom"))}
}

func (l *PromV1Service) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.CreateGroup")
	defer span.End()
	return l.logic.CreateGroup(ctx, req)
}

func (l *PromV1Service) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.UpdateGroup")
	defer span.End()
	return l.logic.UpdateGroup(ctx, req)
}

func (l *PromV1Service) UpdateGroupStatus(ctx context.Context, request *pb.UpdateGroupsStatusRequest) (*pb.UpdateGroupsStatusReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.UpdateGroupStatus")
	defer span.End()
	return l.logic.UpdateGroupsStatus(ctx, request)
}

func (l *PromV1Service) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.DeleteGroup")
	defer span.End()
	return l.logic.DeleteGroup(ctx, req)
}

func (l *PromV1Service) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.GetGroup")
	defer span.End()
	return l.logic.GetGroup(ctx, req)
}

func (l *PromV1Service) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.ListGroup")
	defer span.End()
	return l.logic.ListGroup(ctx, req)
}

func (l *PromV1Service) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.CreateStrategy")
	defer span.End()
	return l.logic.CreateStrategy(ctx, req)
}

func (l *PromV1Service) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.UpdateStrategy")
	defer span.End()
	return l.logic.UpdateStrategy(ctx, req)
}

func (l *PromV1Service) UpdateStrategiesStatus(ctx context.Context, req *pb.UpdateStrategiesStatusRequest) (*pb.UpdateStrategiesStatusReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.UpdateStrategiesStatus")
	defer span.End()
	return l.logic.UpdateStrategiesStatus(ctx, req)
}

func (l *PromV1Service) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.DeleteStrategy")
	defer span.End()
	return l.logic.DeleteStrategy(ctx, req)
}

func (l *PromV1Service) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.GetStrategy")
	defer span.End()
	return l.logic.GetStrategy(ctx, req)
}

func (l *PromV1Service) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromV1Service.ListStrategy")
	defer span.End()
	return l.logic.ListStrategy(ctx, req)
}
