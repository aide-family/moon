package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	INodeLogic interface {
		CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeReply, error)
		UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeReply, error)
		DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeReply, error)
		GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeReply, error)
		ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error)
	}

	NodeService struct {
		pb.UnimplementedNodeServer

		logger *log.Helper
		logic  INodeLogic
	}
)

var _ pb.NodeServer = (*NodeService)(nil)

func NewNodeService(logic INodeLogic, logger log.Logger) *NodeService {
	return &NodeService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Node"))}
}

func (l *NodeService) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "NodeService.CreateNode")
	defer span.End()
	return l.logic.CreateNode(ctx, req)
}

func (l *NodeService) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "NodeService.UpdateNode")
	defer span.End()
	return l.logic.UpdateNode(ctx, req)
}

func (l *NodeService) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "NodeService.DeleteNode")
	defer span.End()
	return l.logic.DeleteNode(ctx, req)
}

func (l *NodeService) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "NodeService.GetNode")
	defer span.End()
	return l.logic.GetNode(ctx, req)
}

func (l *NodeService) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "NodeService.ListNode")
	defer span.End()
	return l.logic.ListNode(ctx, req)
}
