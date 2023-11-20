package promservice

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/strategy/group"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/pkg/model/strategygroup"
	"prometheus-manager/pkg/util/slices"
)

type GroupService struct {
	pb.UnimplementedGroupServer
	strategyGroupBiz *biz.StrategyGroupBiz

	log *log.Helper
}

func NewGroupService(strategyGroupBiz *biz.StrategyGroupBiz, logger log.Logger) *GroupService {
	return &GroupService{
		log:              log.NewHelper(log.With(logger, "module", "service.prom.strategy.group")),
		strategyGroupBiz: strategyGroupBiz,
	}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	strategyGroup := &dobo.StrategyGroupBO{
		Name:        req.GetName(),
		Remark:      req.GetRemark(),
		CategoryIds: req.GetCategoryIds(),
	}
	strategyGroup, err := s.strategyGroupBiz.Create(ctx, strategyGroup)
	if err != nil {
		return nil, err
	}
	return &pb.CreateGroupReply{
		Id: strategyGroup.Id,
	}, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	strategyGroup := &dobo.StrategyGroupBO{
		Id:          req.GetId(),
		Name:        req.GetName(),
		Remark:      req.GetRemark(),
		CategoryIds: req.GetCategoryIds(),
	}
	if _, err := s.strategyGroupBiz.UpdateById(ctx, strategyGroup); err != nil {
		return nil, err
	}
	return &pb.UpdateGroupReply{
		Id: req.GetId(),
	}, nil
}

func (s *GroupService) BatchUpdateGroupStatus(ctx context.Context, req *pb.BatchUpdateGroupStatusRequest) (*pb.BatchUpdateGroupStatusReply, error) {
	if err := s.strategyGroupBiz.BatchUpdateStatus(ctx, req.GetStatus(), req.GetIds()); err != nil {
		return nil, err
	}
	return &pb.BatchUpdateGroupStatusReply{
		Ids: req.GetIds(),
	}, nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	if err := s.strategyGroupBiz.DeleteByIds(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteGroupReply{
		Id: req.GetId(),
	}, nil
}

func (s *GroupService) BatchDeleteGroup(ctx context.Context, req *pb.BatchDeleteGroupRequest) (*pb.BatchDeleteGroupReply, error) {
	if err := s.strategyGroupBiz.DeleteByIds(ctx, req.GetIds()...); err != nil {
		return nil, err
	}
	return &pb.BatchDeleteGroupReply{
		Ids: req.GetIds(),
	}, nil
}

func (s *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	detail, err := s.strategyGroupBiz.GetById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetGroupReply{
		Detail: detail.ToApiPromPromGroup(),
	}, nil
}

func (s *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		strategygroup.Like(req.GetKeyword()),
	}
	list, err := s.strategyGroupBiz.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return &pb.ListGroupReply{
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(list, func(t *dobo.StrategyGroupBO) *api.PromGroup {
			return t.ToApiPromPromGroup()
		}),
	}, nil
}

func (s *GroupService) SelectGroup(ctx context.Context, req *pb.SelectGroupRequest) (*pb.SelectGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	scopes := []query.ScopeMethod{
		strategygroup.Like(req.GetKeyword()),
	}
	selectList, err := s.strategyGroupBiz.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return &pb.SelectGroupReply{
		List: slices.To(selectList, func(t *dobo.StrategyGroupBO) *api.PromGroupSelectV1 {
			return t.ToApiPromGroupSelectV1()
		}),
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
	}, nil
}

func (s *GroupService) ImportGroup(ctx context.Context, req *pb.ImportGroupRequest) (*pb.ImportGroupReply, error) {
	return &pb.ImportGroupReply{}, nil
}

func (s *GroupService) ExportGroup(ctx context.Context, req *pb.ExportGroupRequest) (*pb.ExportGroupReply, error) {
	return &pb.ExportGroupReply{}, nil
}
