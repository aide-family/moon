package strategy

import (
	"context"
	"fmt"
	"strings"

	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	strategyapi "github.com/aide-family/moon/api/admin/strategy"
)

// Service 策略管理服务
type Service struct {
	strategyapi.UnimplementedStrategyServer
	templateBiz      *biz.TemplateBiz
	strategyBiz      *biz.StrategyBiz
	strategyGroupBiz *biz.StrategyGroupBiz
}

// NewStrategyService 创建策略管理服务
func NewStrategyService(templateBiz *biz.TemplateBiz, strategy *biz.StrategyBiz, strategyGroupBiz *biz.StrategyGroupBiz) *Service {
	return &Service{
		templateBiz:      templateBiz,
		strategyBiz:      strategy,
		strategyGroupBiz: strategyGroupBiz,
	}
}

// CreateStrategyGroup 创建策略组
func (s *Service) CreateStrategyGroup(ctx context.Context, req *strategyapi.CreateStrategyGroupRequest) (*strategyapi.CreateStrategyGroupReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	req.TeamId = claims.Team
	params := build.NewBuilder().WithCreateBoStrategyGroup(req).ToCreateStrategyGroupBO()
	if _, err := s.strategyGroupBiz.CreateStrategyGroup(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.CreateStrategyGroupReply{}, nil
}

// DeleteStrategyGroup 删除策略组
func (s *Service) DeleteStrategyGroup(ctx context.Context, req *strategyapi.DeleteStrategyGroupRequest) (*strategyapi.DeleteStrategyGroupReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.DelStrategyGroupParams{
		ID:     req.GetId(),
		TeamID: claims.Team,
	}
	if err := s.strategyGroupBiz.DeleteStrategyGroup(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.DeleteStrategyGroupReply{}, nil
}

// ListStrategyGroup 策略组列表
func (s *Service) ListStrategyGroup(ctx context.Context, req *strategyapi.ListStrategyGroupRequest) (*strategyapi.ListStrategyGroupReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	req.TeamId = claims.Team
	params := build.NewBuilder().WithContext(ctx).WithListStrategyGroup(req).ToListStrategyGroupBO()
	listPage, err := s.strategyGroupBiz.ListPage(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.ListStrategyGroupReply{
		Pagination: build.NewPageBuilder(params.Page).ToAPI(),
		List: types.SliceTo(listPage, func(strategy *bizmodel.StrategyGroup) *admin.StrategyGroupItem {
			return build.NewBuilder().WithContext(ctx).WithAPIStrategyGroup(strategy).ToAPI()
		}),
	}, nil
}

// GetStrategyGroup 获取策略组详情
func (s *Service) GetStrategyGroup(ctx context.Context, req *strategyapi.GetStrategyGroupRequest) (*strategyapi.GetStrategyGroupReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.GetStrategyGroupDetailParams{
		ID:     req.GetId(),
		TeamID: claims.GetTeam(),
	}
	groupDetail, err := s.strategyGroupBiz.GetStrategyGroupDetail(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.GetStrategyGroupReply{Detail: build.NewBuilder().WithContext(ctx).WithAPIStrategyGroup(groupDetail).ToAPI()}, nil
}

// UpdateStrategyGroup 更新策略组
func (s *Service) UpdateStrategyGroup(ctx context.Context, req *strategyapi.UpdateStrategyGroupRequest) (*strategyapi.UpdateStrategyGroupReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	req.TeamId = claims.GetTeam()
	params := build.NewBuilder().WithUpdateBoStrategyGroup(req).ToUpdateStrategyGroupBO()
	if err := s.strategyGroupBiz.UpdateStrategyGroup(ctx, params); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateStrategyGroupReply{}, nil
}

// UpdateStrategyGroupStatus 更新策略组状态
func (s *Service) UpdateStrategyGroupStatus(ctx context.Context, req *strategyapi.UpdateStrategyGroupStatusRequest) (*strategyapi.UpdateStrategyGroupStatusReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	param := &bo.UpdateStrategyGroupStatusParams{
		IDs:    req.GetIds(),
		Status: vobj.Status(req.GetStatus()),
		TeamID: claims.GetTeam(),
	}
	if err := s.strategyGroupBiz.UpdateStatus(ctx, param); err != nil {
		return nil, err
	}
	return &strategyapi.UpdateStrategyGroupStatusReply{}, nil
}

// CreateStrategy 创建策略
func (s *Service) CreateStrategy(ctx context.Context, req *strategyapi.CreateStrategyRequest) (*strategyapi.CreateStrategyReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// 校验数组是否有重复数据
	if has := types.SlicesHasDuplicates(req.GetStrategyLevel(), func(request *strategyapi.CreateStrategyLevelRequest) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d-", request.GetLevelId()))
		return sb.String()
	}); has {
		return nil, merr.ErrorI18nStrategyLevelRepeatErr(ctx)
	}
	req.TeamId = claims.GetTeam()
	param := build.NewBuilder().WithCreateBoStrategy(req).ToCreateStrategyBO()
	if _, err := s.strategyBiz.CreateStrategy(ctx, param); err != nil {
		return nil, err
	}
	return &strategyapi.CreateStrategyReply{}, nil
}

// UpdateStrategy 更新策略
func (s *Service) UpdateStrategy(ctx context.Context, req *strategyapi.UpdateStrategyRequest) (*strategyapi.UpdateStrategyReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	// 校验数组是否有重复数据
	if has := types.SlicesHasDuplicates(req.GetData().GetStrategyLevel(), func(request *strategyapi.CreateStrategyLevelRequest) string {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%d-", request.GetLevelId()))
		return sb.String()
	}); has {
		return nil, merr.ErrorI18nStrategyLevelRepeatErr(ctx)
	}

	req.GetData().TeamId = claims.GetTeam()
	param := build.NewBuilder().WithUpdateBoStrategy(req).ToUpdateStrategyBO()
	if err := s.strategyBiz.UpdateByID(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.UpdateStrategyReply{}, nil
}

// DeleteStrategy 删除策略
func (s *Service) DeleteStrategy(ctx context.Context, req *strategyapi.DeleteStrategyRequest) (*strategyapi.DeleteStrategyReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	param := &bo.DelStrategyParams{
		TeamID: claims.GetTeam(),
		ID:     req.GetId(),
	}
	if err := s.strategyBiz.DeleteByID(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.DeleteStrategyReply{}, nil
}

// GetStrategy 获取策略详情
func (s *Service) GetStrategy(ctx context.Context, req *strategyapi.GetStrategyRequest) (*strategyapi.GetStrategyReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	param := &bo.GetStrategyDetailParams{
		TeamID: claims.GetTeam(),
		ID:     req.GetId(),
	}
	strategy, err := s.strategyBiz.GetStrategy(ctx, param)
	if err != nil {
		return nil, err
	}
	return &strategyapi.GetStrategyReply{
		Detail: build.NewBuilder().WithAPIStrategy(strategy).ToAPI(ctx),
	}, nil
}

// ListStrategy 获取策略列表
func (s *Service) ListStrategy(ctx context.Context, req *strategyapi.ListStrategyRequest) (*strategyapi.ListStrategyReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.QueryStrategyListParams{
		TeamID:     claims.GetTeam(),
		Page:       types.NewPagination(req.GetPagination()),
		Status:     vobj.Status(req.GetStatus()),
		Keyword:    req.GetKeyword(),
		SourceType: vobj.TemplateSourceType(req.GetDatasourceType()),
	}
	strategies, err := s.strategyBiz.StrategyPage(ctx, params)
	if err != nil {
		return nil, err
	}
	return &strategyapi.ListStrategyReply{
		Pagination: build.NewPageBuilder(params.Page).ToAPI(),
		List: types.SliceTo(strategies, func(strategy *bizmodel.Strategy) *admin.StrategyItem {
			return build.NewBuilder().WithAPIStrategy(strategy).ToAPI(ctx)
		}),
	}, nil
}

// UpdateStrategyStatus 更新策略状态
func (s *Service) UpdateStrategyStatus(ctx context.Context, req *strategyapi.UpdateStrategyStatusRequest) (*strategyapi.UpdateStrategyStatusReply, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	params := &bo.UpdateStrategyStatusParams{
		Ids:    req.GetIds(),
		Status: vobj.Status(req.GetStatus()),
		TeamID: claims.GetTeam(),
	}
	err := s.strategyBiz.UpdateStatus(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.UpdateStrategyStatusReply{}, nil
}

// CopyStrategy 复制策略
func (s *Service) CopyStrategy(ctx context.Context, req *strategyapi.CopyStrategyRequest) (*strategyapi.CopyStrategyReply, error) {
	param := &bo.CopyStrategyParams{
		StrategyID: req.StrategyId,
	}
	strategy, err := s.strategyBiz.CopyStrategy(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.CopyStrategyReply{Id: strategy.ID}, nil
}
