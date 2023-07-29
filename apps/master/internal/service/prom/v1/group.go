package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IGroupLogic interface {
		CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error)
		UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error)
		DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error)
		GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error)
		ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error)
	}

	GroupService struct {
		pb.UnimplementedGroupServer

		logger *log.Helper
		logic  IGroupLogic
	}
)

var _ pb.GroupServer = (*GroupService)(nil)

func NewGroupService(logic IGroupLogic, logger log.Logger) *GroupService {
	return &GroupService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Group"))}
}

func (l *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "GroupService.CreateGroup")
	defer span.End()
	return l.logic.CreateGroup(ctx, req)
}

func (l *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "GroupService.UpdateGroup")
	defer span.End()
	return l.logic.UpdateGroup(ctx, req)
}

func (l *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "GroupService.DeleteGroup")
	defer span.End()
	return l.logic.DeleteGroup(ctx, req)
}

func (l *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "GroupService.GetGroup")
	defer span.End()
	return l.logic.GetGroup(ctx, req)
}

func (l *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "GroupService.ListGroup")
	defer span.End()
	return l.logic.ListGroup(ctx, req)
}
