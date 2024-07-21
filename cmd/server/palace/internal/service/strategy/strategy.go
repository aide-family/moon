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

type Service struct {
	strategyapi.UnimplementedStrategyServer
	templateBiz *biz.TemplateBiz
	strategyBiz *biz.StrategyBiz
}

func NewStrategyService(templateBiz *biz.TemplateBiz, strategy *biz.StrategyBiz) *Service {
	return &Service{
		templateBiz: templateBiz,
		strategyBiz: strategy,
	}
}

func (s *Service) CreateStrategyGroup(ctx context.Context, req *strategyapi.CreateStrategyGroupRequest) (*strategyapi.CreateStrategyGroupReply, error) {
	return &strategyapi.CreateStrategyGroupReply{}, nil
}

func (s *Service) DeleteStrategyGroup(ctx context.Context, req *strategyapi.DeleteStrategyGroupRequest) (*strategyapi.DeleteStrategyGroupReply, error) {
	return &strategyapi.DeleteStrategyGroupReply{}, nil
}

func (s *Service) ListStrategyGroup(ctx context.Context, req *strategyapi.ListStrategyGroupRequest) (*strategyapi.ListStrategyGroupReply, error) {
	return &strategyapi.ListStrategyGroupReply{}, nil
}

func (s *Service) GetStrategyGroup(ctx context.Context, req *strategyapi.GetStrategyGroupRequest) (*strategyapi.GetStrategyGroupReply, error) {
	return &strategyapi.GetStrategyGroupReply{}, nil
}

func (s *Service) UpdateStrategyGroup(ctx context.Context, req *strategyapi.UpdateStrategyGroupRequest) (*strategyapi.UpdateStrategyGroupReply, error) {
	return &strategyapi.UpdateStrategyGroupReply{}, nil
}

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

	param := build.NewBuilder().WithCreateBoStrategy(req).ToCreateStrategyBO(claims.GetTeam())

	if _, err := s.strategyBiz.CreateStrategy(ctx, param); err != nil {
		return nil, err
	}
	return &strategyapi.CreateStrategyReply{}, nil
}

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
	param := build.NewBuilder().WithUpdateBoStrategy(req).ToUpdateStrategyBO(claims.GetTeam())
	if err := s.strategyBiz.UpdateByID(ctx, param); !types.IsNil(err) {
		return nil, err
	}
	return &strategyapi.UpdateStrategyReply{}, nil
}

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
		Detail: build.NewBuilder().WithApiStrategy(strategy).ToApi(ctx),
	}, nil
}

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
		Pagination: build.NewPageBuilder(params.Page).ToApi(),
		List: types.SliceTo(strategies, func(strategy *bizmodel.Strategy) *admin.Strategy {
			return build.NewBuilder().WithApiStrategy(strategy).ToApi(ctx)
		}),
	}, nil
}
