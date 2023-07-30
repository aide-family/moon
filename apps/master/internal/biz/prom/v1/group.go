package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	promPB "prometheus-manager/api/prom"
	"prometheus-manager/pkg/times"

	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/dal/model"
)

type (
	IGroupRepo interface {
		V1Repo
		CreateGroup(ctx context.Context, m *model.PromNodeDirFileGroup) error
		UpdateGroupById(ctx context.Context, id uint32, m *model.PromNodeDirFileGroup) error
		DeleteGroupById(ctx context.Context, id uint32) error
		GetGroupById(ctx context.Context, id uint32) (*model.PromNodeDirFileGroup, error)
		ListGroup(ctx context.Context, q *GroupListQueryParams) ([]*model.PromNodeDirFileGroup, int64, error)
	}

	GroupLogic struct {
		logger *log.Helper
		repo   IGroupRepo
	}

	GroupListQueryParams struct {
		Offset  int
		Limit   int
		Keyword string
	}
)

var _ service.IGroupLogic = (*GroupLogic)(nil)

func NewGroupLogic(repo IGroupRepo, logger log.Logger) *GroupLogic {
	return &GroupLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Group"))}
}

func (s *GroupLogic) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "GroupLogic.CreateGroup")
	defer span.End()

	if err := s.repo.CreateGroup(ctx, s.buildGroupModel(req.GetGroup())); err != nil {
		s.logger.Errorf("CreateGroup error: %v", err)
		return nil, err
	}

	return &pb.CreateGroupReply{Response: &api.Response{Message: "add group success"}}, nil
}

func (s *GroupLogic) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "GroupLogic.UpdateGroup")
	defer span.End()

	if err := s.repo.UpdateGroupById(ctx, req.GetId(), s.buildGroupModel(req.GetGroup())); err != nil {
		s.logger.Errorf("UpdateGroup error: %v", err)
		return nil, err
	}

	return &pb.UpdateGroupReply{Response: &api.Response{Message: "update group success"}}, nil
}

func (s *GroupLogic) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "GroupLogic.DeleteGroup")
	defer span.End()

	if err := s.repo.DeleteGroupById(ctx, req.GetId()); err != nil {
		s.logger.Errorf("DeleteGroup error: %v", err)
		return nil, err
	}

	return &pb.DeleteGroupReply{Response: &api.Response{Message: "delete group success"}}, nil
}

func (s *GroupLogic) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "GroupLogic.GetGroup")
	defer span.End()

	groupInfo, err := s.repo.GetGroupById(ctx, req.GetId())
	if err != nil {
		s.logger.Errorf("GetGroup error: %v", err)
		return nil, err
	}

	return &pb.GetGroupReply{Response: &api.Response{Message: "get group success"}, Group: s.buildGroupItem(groupInfo)}, nil
}

func (s *GroupLogic) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "GroupLogic.ListGroup")
	defer span.End()

	limit := int(req.GetPage().GetSize())
	offset := int(req.GetPage().GetCurrent()-1) * limit

	query := &GroupListQueryParams{
		Offset:  offset,
		Limit:   limit,
		Keyword: req.GetParams().GetKeyword(),
	}

	groupList, total, err := s.repo.ListGroup(ctx, query)
	if err != nil {
		s.logger.Errorf("ListGroup error: %v", err)
		return nil, err
	}

	list := make([]*promPB.GroupItem, 0, len(groupList))
	for _, item := range groupList {
		list = append(list, s.buildGroupItem(item))
	}

	return &pb.ListGroupReply{
		Response: &api.Response{Message: "get group list success"},
		Page: &api.PageReply{
			Current: req.GetPage().GetCurrent(),
			Size:    req.GetPage().GetSize(),
			Total:   total,
		}, List: list,
	}, nil
}

func (s *GroupLogic) buildGroupModel(req *promPB.GroupItem) *model.PromNodeDirFileGroup {
	if req == nil {
		return nil
	}

	return &model.PromNodeDirFileGroup{
		Name:   req.GetName(),
		Remark: req.GetRemark(),
		FileID: int32(req.GetFileId()),
	}
}

func (s *GroupLogic) buildGroupItem(m *model.PromNodeDirFileGroup) *promPB.GroupItem {
	if m == nil {
		return nil
	}

	return &promPB.GroupItem{
		Id:        uint32(m.ID),
		Name:      m.Name,
		Remark:    m.Remark,
		FileId:    uint32(m.FileID),
		CreatedAt: times.TimeToUnix(m.CreatedAt),
		UpdatedAt: times.TimeToUnix(m.UpdatedAt),
	}
}
