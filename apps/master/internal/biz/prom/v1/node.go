package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
)

type (
	INodeRepo interface {
		V1Repo
	}

	NodeLogic struct {
		logger *log.Helper
		repo   INodeRepo
	}
)

var _ service.INodeLogic = (*NodeLogic)(nil)

func NewNodeLogic(repo INodeRepo, logger log.Logger) *NodeLogic {
	return &NodeLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Node"))}
}

func (s *NodeLogic) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "NodeLogic.CreateNode")
	defer span.End()
	return nil, nil
}
func (s *NodeLogic) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "NodeLogic.UpdateNode")
	defer span.End()
	return nil, nil
}
func (s *NodeLogic) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "NodeLogic.DeleteNode")
	defer span.End()
	return nil, nil
}
func (s *NodeLogic) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "NodeLogic.GetNode")
	defer span.End()
	return nil, nil
}
func (s *NodeLogic) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "NodeLogic.ListNode")
	defer span.End()
	return nil, nil
}
