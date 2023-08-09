package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IPromLogic interface {
		CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error)
		UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error)
		DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error)
		GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error)
		ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error)
		CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error)
		UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error)
		DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error)
		GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error)
		ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error)
	}

	PromService struct {
		pb.UnimplementedPromServer

		logger *log.Helper
		logic  IPromLogic
	}
)

var _ pb.PromServer = (*PromService)(nil)

func NewPromService(logic IPromLogic, logger log.Logger) *PromService {
	return &PromService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Prom"))}
}

func (l *PromService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.CreateGroup")
	defer span.End()
	return l.logic.CreateGroup(ctx, req)
}

func (l *PromService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.UpdateGroup")
	defer span.End()
	return l.logic.UpdateGroup(ctx, req)
}

func (l *PromService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.DeleteGroup")
	defer span.End()
	return l.logic.DeleteGroup(ctx, req)
}

func (l *PromService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.GetGroup")
	defer span.End()
	return l.logic.GetGroup(ctx, req)
}

func (l *PromService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.ListGroup")
	defer span.End()
	return l.logic.ListGroup(ctx, req)
}

func (l *PromService) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.CreateStrategy")
	defer span.End()
	return l.logic.CreateStrategy(ctx, req)
}

func (l *PromService) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.UpdateStrategy")
	defer span.End()
	return l.logic.UpdateStrategy(ctx, req)
}

func (l *PromService) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.DeleteStrategy")
	defer span.End()
	return l.logic.DeleteStrategy(ctx, req)
}

func (l *PromService) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.GetStrategy")
	defer span.End()
	return l.logic.GetStrategy(ctx, req)
}

func (l *PromService) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "PromService.ListStrategy")
	defer span.End()
	return l.logic.ListStrategy(ctx, req)
}
