package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	palace "github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
)

type TeamStrategyMetricService struct {
	palace.UnimplementedTeamStrategyMetricServer

	teamStrategyBiz       *biz.TeamStrategy
	teamStrategyMetricBiz *biz.TeamStrategyMetric
}

func NewTeamStrategyMetricService(teamStrategyBiz *biz.TeamStrategy, teamStrategyMetricBiz *biz.TeamStrategyMetric) *TeamStrategyMetricService {
	return &TeamStrategyMetricService{
		teamStrategyBiz:       teamStrategyBiz,
		teamStrategyMetricBiz: teamStrategyMetricBiz,
	}
}

func (s *TeamStrategyMetricService) SaveTeamMetricStrategy(ctx context.Context, req *palace.SaveTeamMetricStrategyRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToSaveTeamMetricStrategyParams(req)
	if err := s.teamStrategyMetricBiz.SaveTeamMetricStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamStrategyMetricService) DeleteTeamMetricStrategy(ctx context.Context, req *palace.DeleteTeamMetricStrategyRequest) (*palacecommon.EmptyReply, error) {
	if err := s.teamStrategyBiz.DeleteTeamStrategy(ctx, req.GetStrategyId()); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamStrategyMetricService) TeamMetricStrategyDetail(ctx context.Context, req *palace.TeamMetricStrategyDetailRequest) (*palacecommon.TeamStrategyMetricItem, error) {
	strategyMetricDo, err := s.teamStrategyMetricBiz.GetTeamMetricStrategyByStrategyID(ctx, req.GetStrategyId())
	if err != nil {
		return nil, err
	}
	return build.ToTeamMetricStrategyItem(strategyMetricDo), nil
}

func (s *TeamStrategyMetricService) SaveTeamMetricStrategyLevel(ctx context.Context, req *palace.SaveTeamMetricStrategyLevelRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToSaveTeamMetricStrategyLevelParams(req)
	if err := s.teamStrategyMetricBiz.SaveTeamMetricStrategyLevel(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamStrategyMetricService) DeleteTeamMetricStrategyLevel(ctx context.Context, req *palace.DeleteTeamMetricStrategyLevelRequest) (*palacecommon.EmptyReply, error) {
	if err := s.teamStrategyMetricBiz.DeleteTeamMetricStrategyLevels(ctx, req.GetStrategyMetricLevelIds()); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}

func (s *TeamStrategyMetricService) TeamMetricStrategyLevelDetail(ctx context.Context, req *palace.TeamMetricStrategyLevelDetailRequest) (*palacecommon.TeamStrategyMetricLevelItem, error) {
	strategyMetricLevelDo, err := s.teamStrategyMetricBiz.GetTeamMetricStrategyLevel(ctx, req.GetStrategyMetricLevelId())
	if err != nil {
		return nil, err
	}
	return build.ToTeamMetricStrategyRuleItem(strategyMetricLevelDo), nil
}

func (s *TeamStrategyMetricService) TeamMetricStrategyLevelList(ctx context.Context, req *palace.TeamMetricStrategyLevelListRequest) (*palace.TeamMetricStrategyLevelListReply, error) {
	params := build.ToListTeamMetricStrategyLevelsParams(req)
	strategyMetricLevelBo, err := s.teamStrategyMetricBiz.ListTeamMetricStrategyLevels(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.TeamMetricStrategyLevelListReply{
		Items:      build.ToTeamMetricStrategyRuleItems(strategyMetricLevelBo.Items),
		Pagination: build.ToPaginationReply(strategyMetricLevelBo.PaginationReply),
	}, nil
}

func (s *TeamStrategyMetricService) UpdateTeamMetricStrategyLevelStatus(ctx context.Context, req *palace.UpdateTeamMetricStrategyLevelStatusRequest) (*palacecommon.EmptyReply, error) {
	params := build.ToUpdateTeamMetricStrategyLevelStatusParams(req)
	if err := s.teamStrategyMetricBiz.UpdateTeamMetricStrategyLevelStatus(ctx, params); err != nil {
		return nil, err
	}
	return &palacecommon.EmptyReply{}, nil
}
