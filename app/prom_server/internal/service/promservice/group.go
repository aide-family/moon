package promservice

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/strategy"

	"prometheus-manager/api"
	pb "prometheus-manager/api/server/prom/strategy/group"
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
	if err := s.strategyGroupBiz.BatchUpdateStatus(ctx, vo.Status(req.GetStatus()), req.GetIds()); err != nil {
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
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	list, err := s.strategyGroupBiz.List(ctx, &bo.ListGroupReq{
		Page:              pgInfo,
		Keyword:           req.GetKeyword(),
		Status:            vo.Status(req.GetStatus()),
		PreloadCategories: true,
		Ids:               req.GetIds(),
	})
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

func (s *GroupService) ListAllGroupDetail(ctx context.Context, req *pb.ListAllGroupDetailRequest) (*pb.ListAllGroupDetailReply, error) {
	list := make([]*api.GroupSimple, 0)
	wheres := []basescopes.ScopeMethod{
		basescopes.StatusEQ(vo.StatusEnabled),
		func(db *gorm.DB) *gorm.DB {
			return db.Preload(do.PromGroupPreloadFieldPromStrategies, basescopes.StatusEQ(vo.StatusEnabled))
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Preload(strings.Join([]string{
				do.PromGroupPreloadFieldPromStrategies,
				do.PromStrategyPreloadFieldEndpoint}, "."), basescopes.StatusEQ(vo.StatusEnabled))
		},
	}

	defaultId := uint32(0)
	if len(req.GetGroupIds()) > 0 {
		wheres = append(wheres, basescopes.InIds(req.GetGroupIds()...))
		ids := req.GetGroupIds()
		// 排序
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})
		defaultId = ids[0] - 1
	}

	for {
		strategyGroupBOS, err := s.strategyGroupBiz.ListAllLimit(ctx, 1000, append(wheres, basescopes.IdGT(defaultId))...)
		if err != nil {
			s.log.Errorf("ListAllGroupDetail error: %v", err)
			break
		}
		if len(strategyGroupBOS) == 0 {
			break
		}
		list = append(list, slices.ToFilter(strategyGroupBOS, func(t *bo.StrategyGroupBO) (*api.GroupSimple, bool) {
			if t.Status != vo.StatusEnabled {
				return nil, false
			}
			t.PromStrategies = slices.ToFilter(t.GetPromStrategies(), func(ru *bo.StrategyBO) (*bo.StrategyBO, bool) {
				if ru == nil || ru.Status != vo.StatusEnabled || ru.Endpoint == nil || ru.Endpoint.Endpoint == "" {
					return nil, false
				}
				return ru, true
			})
			return t.ToSimpleApi(), true
		})...)
		defaultId = strategyGroupBOS[len(strategyGroupBOS)-1].Id
	}
	return &pb.ListAllGroupDetailReply{
		GroupList: list,
	}, nil
}

func (s *GroupService) SelectGroup(ctx context.Context, req *pb.SelectGroupRequest) (*pb.SelectGroupReply, error) {
	pgReq := req.GetPage()
	pgInfo := bo.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	selectList, err := s.strategyGroupBiz.List(ctx, &bo.ListGroupReq{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
		Status:  vo.Status(req.GetStatus()),
	})
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
	if len(req.GetGroups()) == 0 {
		return &pb.ImportGroupReply{}, nil
	}
	groupNames := slices.To(req.GetGroups(), func(t *api.PromRuleGroup) string {
		return t.GetName()
	})

	promStrategyGroups, err := s.strategyGroupBiz.ListAllLimit(ctx, len(groupNames), basescopes.NameIn(groupNames...))
	if err != nil {
		return nil, err
	}
	promStrategyGroupNameMap := make(map[string]*bo.StrategyGroupBO)
	for _, promStrategyGroup := range promStrategyGroups {
		promStrategyGroupNameMap[promStrategyGroup.Name] = promStrategyGroup
	}

	// 构建创建的数据
	newPromStrategyGroups := slices.To(req.GetGroups(), func(t *api.PromRuleGroup) *bo.StrategyGroupBO {
		strategyGroup := &bo.StrategyGroupBO{
			Name: t.GetName(),
		}
		if promStrategyGroup, ok := promStrategyGroupNameMap[t.GetName()]; ok {
			strategyGroup = promStrategyGroup
		}
		strategyGroup.PromStrategies = slices.To(t.GetRules(), func(rule *api.PromRule) *bo.StrategyBO {
			labels := rule.GetLabels()
			annotations := rule.GetAnnotations()
			levelId := uint64(req.GetDefaultLevel())
			levelIdStr, ok := labels[strategy.MetricLevelId]
			if ok {
				levelId, err = strconv.ParseUint(levelIdStr, 10, 64)
				if err != nil {
					levelId = uint64(req.GetDefaultLevel())
				}
			}

			strategyDetail := &bo.StrategyBO{
				Alert:        rule.GetAlert(),
				Expr:         rule.GetExpr(),
				Duration:     rule.GetFor(),
				Labels:       (*strategy.Labels)(&labels),
				Annotations:  (*strategy.Annotations)(&annotations),
				Status:       vo.StatusDisabled,
				Remark:       rule.GetAlert(),
				GroupId:      strategyGroup.Id,
				AlarmLevelId: uint32(levelId),
				AlarmPages:   slices.To(req.GetDefaultAlarmPageIds(), func(id uint32) *bo.AlarmPageBO { return &bo.AlarmPageBO{Id: id} }),
				Categories:   slices.To(req.GetDefaultCategoryIds(), func(id uint32) *bo.DictBO { return &bo.DictBO{Id: id} }),
				PromNotifies: slices.To(req.GetDefaultAlarmNotifyIds(), func(id uint32) *bo.NotifyBO { return &bo.NotifyBO{Id: id} }),
				EndpointId:   req.GetDatasourceId(),
			}
			return strategyDetail
		})
		return strategyGroup
	})
	s.log.Infow("newPromStrategyGroups", newPromStrategyGroups)
	// 执行导入创建
	newPromStrategyGroups, err = s.strategyGroupBiz.BatchCreate(ctx, newPromStrategyGroups)
	if err != nil {
		return nil, err
	}

	// 根据规则组名称查询是否存在
	return &pb.ImportGroupReply{
		Ids: slices.To(newPromStrategyGroups, func(t *bo.StrategyGroupBO) uint32 { return t.Id }),
	}, nil
}

func (s *GroupService) ExportGroup(ctx context.Context, req *pb.ExportGroupRequest) (*pb.ExportGroupReply, error) {
	return &pb.ExportGroupReply{}, nil
}
