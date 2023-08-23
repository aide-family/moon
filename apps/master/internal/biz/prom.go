package biz

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"

	"prometheus-manager/apps/master/internal/service"
	"prometheus-manager/pkg/dal/model"
)

type (
	IPromV1Repo interface {
		V1Repo
		CreateGroup(ctx context.Context, m *model.PromGroup) error
		UpdateGroupByID(ctx context.Context, id int32, m *model.PromGroup) error
		UpdateGroupsStatusByIds(ctx context.Context, ids []int32, status prom.Status) error
		DeleteGroupByID(ctx context.Context, id int32) error
		GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error)
		Groups(ctx context.Context, req *pb.ListGroupRequest) ([]*model.PromGroup, int64, error)
		SimpleGroups(ctx context.Context, req *pb.ListSimpleGroupRequest) ([]*model.PromGroup, int64, error)

		CreateStrategy(ctx context.Context, m *model.PromStrategy) error
		UpdateStrategyByID(ctx context.Context, id int32, m *model.PromStrategy) error
		UpdateStrategiesStatusByIds(ctx context.Context, ids []int32, status prom.Status) error
		DeleteStrategyByID(ctx context.Context, id int32) error
		StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error)
		Strategies(ctx context.Context, req *pb.ListStrategyRequest) ([]*model.PromStrategy, int64, error)
		GetStrategyByName(ctx context.Context, groupID int32, name string) (*model.PromStrategy, error)
	}

	PromLogic struct {
		logger *log.Helper
		v1Repo IPromV1Repo
	}
)

var _ service.IPromV1Logic = (*PromLogic)(nil)

// NewPromLogic 初始化biz.PromLogic
func NewPromLogic(v1Repo IPromV1Repo, logger log.Logger) *PromLogic {
	return &PromLogic{v1Repo: v1Repo, logger: log.NewHelper(log.With(logger, "module", promModuleName))}
}

// ListSimpleGroup 简单的规则组列表, 用于前端选择用
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) ListSimpleGroup(ctx context.Context, req *pb.ListSimpleGroupRequest) (*pb.ListSimpleGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.ListSimpleGroup")
	defer span.End()

	simpleList, total, err := s.v1Repo.SimpleGroups(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("SimpleGroups err: %v", err)
		return nil, err
	}

	list := make([]*prom.SimpleItem, 0, len(simpleList))
	for _, group := range simpleList {
		list = append(list, &prom.SimpleItem{
			Id:   group.ID,
			Name: group.Name,
		})
	}

	return &pb.ListSimpleGroupReply{
		Groups: list,
		Page: &api.PageReply{
			Current: req.GetPage().GetCurrent(),
			Size:    req.GetPage().GetSize(),
			Total:   total,
		},
		Response: &api.Response{Message: "获取成功"},
	}, nil
}

// CreateGroup 创建Prometheus分组
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.CreateGroup")
	defer span.End()

	insertModel := buildModelPromGroup(req.GetGroup())
	if err := s.v1Repo.CreateGroup(ctx, insertModel); err != nil {
		s.logger.WithContext(ctx).Errorw("创建Prometheus分组失败", insertModel, "err", err)
		return nil, perrors.ErrorLogicCreatePrometheusGroupFailed("创建Prometheus分组失败")
	}

	return &pb.CreateGroupReply{Response: &api.Response{Message: "创建Prometheus group成功"}}, nil
}

// UpdateGroup 更新Prometheus分组
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.UpdateGroup")
	defer span.End()

	edieModel := buildModelPromGroup(req.GetGroup())
	if err := s.v1Repo.UpdateGroupByID(ctx, req.GetId(), edieModel); err != nil {
		s.logger.WithContext(ctx).Errorw("更新Prometheus分组失败", edieModel, "err", err)
		return nil, perrors.ErrorLogicEditPrometheusGroupFailed("更新Prometheus分组失败")
	}

	return &pb.UpdateGroupReply{Response: &api.Response{Message: "更新Prometheus group成功"}}, nil
}

// UpdateGroupsStatus 批量更新Prometheus分组状态
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) UpdateGroupsStatus(ctx context.Context, req *pb.UpdateGroupsStatusRequest) (*pb.UpdateGroupsStatusReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.UpdateGrouStatus")
	defer span.End()

	if err := s.v1Repo.UpdateGroupsStatusByIds(ctx, req.GetIds(), req.GetStatus()); err != nil {
		s.logger.WithContext(ctx).Errorw("更新Prometheus分组状态失败", "ids", req.GetIds(), "status", req.GetStatus(), "err", err)
		return nil, perrors.ErrorLogicEditPrometheusGroupFailed("更新Prometheus分组状态失败").WithCause(err)
	}

	return &pb.UpdateGroupsStatusReply{Response: &api.Response{Message: "更新Prometheus group状态成功"}}, nil
}

// DeleteGroup 删除Prometheus分组
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.DeleteGroup")
	defer span.End()

	if err := s.v1Repo.DeleteGroupByID(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorw("删除Prometheus分组失败", "id", req.GetId(), "err", err)
		return nil, perrors.ErrorLogicDeletePrometheusGroupFailed("删除Prometheus分组失败").WithCause(err)
	}

	return &pb.DeleteGroupReply{Response: &api.Response{Message: "删除Prometheus group成功"}}, nil
}

// GetGroup 获取Prometheus分组信息
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.GetGroup")
	defer span.End()

	groupDetail, err := s.v1Repo.GroupDetail(ctx, req.GetId())
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus分组失败", "id", req.GetId(), "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus分组失败")
	}

	return &pb.GetGroupReply{
		Response: &api.Response{Message: "获取Prometheus group成功"},
		Group:    buildGroupItem(groupDetail),
	}, nil
}

// ListGroup 获取Prometheus分组列表
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.ListGroup")
	defer span.End()

	groups, total, err := s.v1Repo.Groups(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus分组列表失败", req, "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus分组列表失败")
	}

	list := make([]*prom.GroupItem, 0, len(groups))
	for _, group := range groups {
		list = append(list, buildGroupItem(group))
	}

	return &pb.ListGroupReply{
		Response: &api.Response{Message: "获取Prometheus group列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		Groups: list,
	}, nil
}

// CreateStrategy 创建Prometheus策略
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.CreateStrategy")
	defer span.End()

	firstStrategyInfo, err := s.v1Repo.GetStrategyByName(ctx, req.GetStrategy().GetGroupId(), req.GetStrategy().Alert)
	if err != nil && !perrors.IsLogicDataNotFound(err) {
		s.logger.WithContext(ctx).Errorw("创建Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}
	if firstStrategyInfo != nil && firstStrategyInfo.ID != 0 {
		return nil, perrors.ErrorLogicDataDuplicate("策略名称已存在").WithMetadata(map[string]string{
			"alert":    req.GetStrategy().GetAlert(),
			"gourp_id": strconv.Itoa(int(req.GetStrategy().GetGroupId())),
		})
	}

	if err = s.v1Repo.CreateStrategy(ctx, buildModelPromStrategy(req.GetStrategy())); err != nil {
		s.logger.WithContext(ctx).Errorw("创建Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}

	return &pb.CreateStrategyReply{Response: &api.Response{Message: "创建Prometheus strategy成功"}}, nil
}

// UpdateStrategiesStatus 批量更新Prometheus策略状态
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) UpdateStrategiesStatus(ctx context.Context, req *pb.UpdateStrategiesStatusRequest) (*pb.UpdateStrategiesStatusReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.UpdateStrategiesStatus")
	defer span.End()

	if err := s.v1Repo.UpdateStrategiesStatusByIds(ctx, req.GetIds(), req.GetStatus()); err != nil {
		s.logger.WithContext(ctx).Errorw("更新Prometheus策略状态失败", "ids", req.GetIds(), "status", req.GetStatus(), "err", err)
		return nil, perrors.ErrorLogicEditPrometheusStrategyFailed("更新Prometheus策略状态失败").WithCause(err)
	}

	return &pb.UpdateStrategiesStatusReply{Response: &api.Response{Message: "更新Prometheus策略状态成功"}}, nil
}

// UpdateStrategy 更新Prometheus策略
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.UpdateStrategy")
	defer span.End()

	firstStrategyInfo, err := s.v1Repo.GetStrategyByName(ctx, req.GetStrategy().GetGroupId(), req.GetStrategy().Alert)
	if err != nil && !perrors.IsLogicDataNotFound(err) {
		s.logger.WithContext(ctx).Errorw("修改Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}
	if firstStrategyInfo != nil && firstStrategyInfo.ID != req.GetId() {
		return nil, perrors.ErrorLogicDataDuplicate("策略名称已存在").WithMetadata(map[string]string{
			"alert":    req.GetStrategy().GetAlert(),
			"group_id": strconv.Itoa(int(req.GetStrategy().GetGroupId())),
		})
	}

	if err = s.v1Repo.UpdateStrategyByID(ctx, req.GetId(), buildModelPromStrategy(req.GetStrategy())); err != nil {
		s.logger.WithContext(ctx).Errorw("修改Prometheus策略失败", req.GetStrategy(), "err", err)
		return nil, err
	}

	return &pb.UpdateStrategyReply{Response: &api.Response{Message: "修改Prometheus strategy成功"}}, nil
}

// DeleteStrategy 删除Prometheus策略
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.DeleteStrategy")
	defer span.End()

	if err := s.v1Repo.DeleteStrategyByID(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorw("删除Prometheus策略失败", "id", req.GetId(), "err", err)
		return nil, err
	}
	return &pb.DeleteStrategyReply{Response: &api.Response{Message: "删除Prometheus策略成功"}}, nil
}

// GetStrategy 获取Prometheus策略
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.GetStrategy")
	defer span.End()

	strategyDetail, err := s.v1Repo.StrategyDetail(ctx, req.GetId())
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus策略失败", "id", req.GetId(), "err", err)
		return nil, err
	}

	return &pb.GetStrategyReply{
		Response: &api.Response{Message: "获取Prometheus策略成功"},
		Strategy: buildStrategyItem(strategyDetail),
	}, nil
}

// ListStrategy 获取Prometheus策略列表
//
//	ctx: 上下文
//	req: 请求参数
func (s *PromLogic) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	ctx, span := otel.Tracer(promModuleName).Start(ctx, "PromLogic.ListStrategy")
	defer span.End()

	strategies, total, err := s.v1Repo.Strategies(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorw("获取Prometheus策略列表失败", req, "err", err)
		return nil, perrors.ErrorServerDatabaseError("获取Prometheus策略列表失败")
	}

	return &pb.ListStrategyReply{
		Response: &api.Response{Message: "获取Prometheus策略列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		Strategies: buildPromStrategies(strategies),
	}, nil
}
