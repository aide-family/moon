package biz

import (
	"context"
	"prometheus-manager/pkg/times"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	promPB "prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/dal/model"
)

type (
	INodeRepo interface {
		V1Repo
		CreateNode(ctx context.Context, m *model.PromNode) error
		UpdateNodeById(ctx context.Context, id uint32, m *model.PromNode) error
		DeleteNodeById(ctx context.Context, id uint32) error
		GetNodeById(ctx context.Context, id uint32) (*model.PromNode, error)
		ListNode(ctx context.Context, q *NodeListQueryParams) ([]*model.PromNode, int64, error)
	}

	NodeLogic struct {
		logger *log.Helper
		repo   INodeRepo
	}

	NodeListQueryParams struct {
		Offset  int
		Limit   int
		Keyword string
	}
)

var _ service.INodeLogic = (*NodeLogic)(nil)

func NewNodeLogic(repo INodeRepo, logger log.Logger) *NodeLogic {
	return &NodeLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Node"))}
}

func (s *NodeLogic) CreateNode(ctx context.Context, req *pb.CreateNodeRequest) (*pb.CreateNodeReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "NodeLogic.CreateNode")
	defer span.End()

	if err := s.repo.CreateNode(ctx, s.buildNodeModel(req.GetNode())); err != nil {
		s.logger.Errorf("CreateNode error: %v", err)
		return nil, err
	}

	return &pb.CreateNodeReply{Response: &api.Response{
		Code:     0,
		Message:  "add node success",
		Metadata: nil,
		Data:     nil,
	}}, nil
}

func (s *NodeLogic) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "NodeLogic.UpdateNode")
	defer span.End()

	if err := s.repo.UpdateNodeById(ctx, req.GetId(), s.buildNodeModel(req.GetNode())); err != nil {
		s.logger.Errorf("UpdateNode error: %v", err)
		return nil, err
	}

	return &pb.UpdateNodeReply{Response: &api.Response{
		Code:     0,
		Message:  "edit node success",
		Metadata: nil,
		Data:     nil,
	}}, nil
}

func (s *NodeLogic) DeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.DeleteNodeReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "NodeLogic.DeleteNode")
	defer span.End()

	if err := s.repo.DeleteNodeById(ctx, req.GetId()); err != nil {
		s.logger.Errorf("DeleteNode error: %v", err)
		return nil, err
	}

	return &pb.DeleteNodeReply{Response: &api.Response{
		Code:     0,
		Message:  "delete node success",
		Metadata: nil,
		Data:     nil,
	}}, nil
}

func (s *NodeLogic) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "NodeLogic.GetNode")
	defer span.End()

	node, err := s.repo.GetNodeById(ctx, req.GetId())
	if err != nil {
		s.logger.Errorf("GetNode error: %v", err)
		return nil, err
	}

	return &pb.GetNodeReply{
		Response: &api.Response{
			Code:     0,
			Message:  "get node success",
			Metadata: nil,
			Data:     nil,
		},
		Node: s.buildNodeItem(node),
	}, nil
}

func (s *NodeLogic) ListNode(ctx context.Context, req *pb.ListNodeRequest) (*pb.ListNodeReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "NodeLogic.ListNode")
	defer span.End()

	limit := int(req.GetPage().GetSize())
	offset := (int(req.GetPage().GetCurrent()) - 1) * limit

	query := &NodeListQueryParams{
		Offset:  offset,
		Limit:   limit,
		Keyword: req.GetParams().GetKeyword(),
	}

	nodes, total, err := s.repo.ListNode(ctx, query)
	if err != nil {
		s.logger.Errorf("ListNode error: %v", err)
		return nil, err
	}

	return &pb.ListNodeReply{
		Response: &api.Response{
			Code:     0,
			Message:  "get node success",
			Metadata: nil,
			Data:     nil,
		},
		List: func() []*promPB.NodeItem {
			list := make([]*promPB.NodeItem, 0, len(nodes))
			for _, node := range nodes {
				list = append(list, s.buildNodeItem(node))
			}
			return list
		}(),
		Page: &api.PageReply{
			Current: req.GetPage().GetCurrent(),
			Size:    req.GetPage().GetSize(),
			Total:   total,
		},
	}, nil
}

func (s *NodeLogic) buildNodeModel(node *promPB.NodeItem) *model.PromNode {
	if node == nil {
		return nil
	}
	return &model.PromNode{
		EnName:     node.GetEnName(),
		ChName:     node.GetCnName(),
		Datasource: node.GetDatasource(),
		Remark:     node.GetRemark(),
	}
}

func (s *NodeLogic) buildNodeItem(node *model.PromNode) *promPB.NodeItem {
	if node == nil {
		return nil
	}
	return &promPB.NodeItem{
		Id:         uint32(node.ID),
		EnName:     node.EnName,
		CnName:     node.ChName,
		Datasource: node.Datasource,
		Remark:     node.Remark,
		CreatedAt:  times.TimeToUnix(node.CreatedAt),
		UpdatedAt:  times.TimeToUnix(node.UpdatedAt),
	}
}
