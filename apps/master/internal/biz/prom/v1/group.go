package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
)

type (
	IGroupRepo interface {
		V1Repo
	}

	GroupLogic struct {
		logger *log.Helper
		repo   IGroupRepo
	}
)

var _ service.IGroupLogic = (*GroupLogic)(nil)

func NewGroupLogic(repo IGroupRepo, logger log.Logger) *GroupLogic {
	return &GroupLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Group"))}
}

func (s *GroupLogic) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "GroupLogic.CreateGroup")
	defer span.End()
	return nil, nil
}
func (s *GroupLogic) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "GroupLogic.UpdateGroup")
	defer span.End()
	return nil, nil
}
func (s *GroupLogic) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "GroupLogic.DeleteGroup")
	defer span.End()
	return nil, nil
}
func (s *GroupLogic) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "GroupLogic.GetGroup")
	defer span.End()
	return nil, nil
}
func (s *GroupLogic) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "GroupLogic.ListGroup")
	defer span.End()
	return nil, nil
}
