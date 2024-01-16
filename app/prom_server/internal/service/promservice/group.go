package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/strategy/group"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
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
	strategyGroup := &bo.StrategyGroupBO{
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
	strategyGroup := &bo.StrategyGroupBO{
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
		Detail: detail.ToApiV1(),
	}, nil
}

func (s *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
		basescopes.StrategyTablePreloadCategories,
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
	}
	list, err := s.strategyGroupBiz.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return &pb.ListGroupReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(list, func(t *bo.StrategyGroupBO) *api.PromGroup {
			return t.ToApiV1()
		}),
	}, nil
}

func (s *GroupService) ListAllGroupDetail(ctx context.Context, _ *pb.ListAllGroupDetailRequest) (*pb.ListAllGroupDetailReply, error) {
	list := make([]*api.PromGroup, 0)
	wheres := []basescopes.ScopeMethod{
		basescopes.StatusEQ(vo.StatusEnabled),
		basescopes.PreloadStrategyGroupPromStrategies(
			basescopes.PreloadKeyEndpoint,
		),
	}
	defaultId := uint32(0)
	for {
		strategyGroupBOS, err := s.strategyGroupBiz.ListAllLimit(ctx, 1000, append(wheres, basescopes.IdGT(defaultId))...)
		if err != nil {
			s.log.Errorf("ListAllGroupDetail error: %v", err)
			break
		}
		if len(strategyGroupBOS) == 0 {
			break
		}
		list = append(list, slices.To(strategyGroupBOS, func(t *bo.StrategyGroupBO) *api.PromGroup {
			return t.ToApiV1()
		})...)
		defaultId = strategyGroupBOS[len(strategyGroupBOS)-1].Id
	}
	return &pb.ListAllGroupDetailReply{
		List: list,
	}, nil
}

func (s *GroupService) SelectGroup(ctx context.Context, req *pb.SelectGroupRequest) (*pb.SelectGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.GetKeyword()),
		basescopes.UpdateAtDesc(),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
	}
	selectList, err := s.strategyGroupBiz.List(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	return &pb.SelectGroupReply{
		List: slices.To(selectList, func(t *bo.StrategyGroupBO) *api.PromGroupSelectV1 {
			return t.ToApiSelectV1()
		}),
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
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
