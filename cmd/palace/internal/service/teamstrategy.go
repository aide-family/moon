package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
)

func NewTeamStrategyService(
	teamStrategyGroupBiz *biz.TeamStrategyGroup,
	teamStrategyBiz *biz.TeamStrategy,
) *TeamStrategyService {
	return &TeamStrategyService{
		teamStrategyGroupBiz: teamStrategyGroupBiz,
		teamStrategyBiz:      teamStrategyBiz,
	}
}

type TeamStrategyService struct {
	palace.UnimplementedTeamStrategyServer
	teamStrategyGroupBiz *biz.TeamStrategyGroup
	teamStrategyBiz      *biz.TeamStrategy
}

func (t *TeamStrategyService) SaveTeamStrategyGroup(ctx context.Context, request *palace.SaveTeamStrategyGroupRequest) (*common.EmptyReply, error) {
	params := &bo.SaveTeamStrategyGroupParams{
		ID:     request.GetGroupId(),
		Name:   request.GetName(),
		Remark: request.GetRemark(),
	}
	if err := t.teamStrategyGroupBiz.SaveTeamStrategyGroup(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) UpdateTeamStrategyGroupStatus(ctx context.Context, request *palace.UpdateTeamStrategyGroupStatusRequest) (*common.EmptyReply, error) {
	params := &bo.UpdateTeamStrategyGroupStatusParams{
		ID:     request.GetGroupId(),
		Status: vobj.GlobalStatus(request.GetStatus()),
	}
	if err := t.teamStrategyGroupBiz.UpdateTeamStrategyGroupStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) DeleteTeamStrategyGroup(ctx context.Context, request *palace.DeleteTeamStrategyGroupRequest) (*common.EmptyReply, error) {
	if err := t.teamStrategyGroupBiz.DeleteTeamStrategyGroup(ctx, request.GetGroupId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) GetTeamStrategyGroup(ctx context.Context, request *palace.GetTeamStrategyGroupRequest) (*common.TeamStrategyGroupItem, error) {
	group, err := t.teamStrategyGroupBiz.GetTeamStrategyGroup(ctx, request.GetGroupId())
	if err != nil {
		return nil, err
	}
	return build.ToTeamStrategyGroupItem(group), nil
}

func (t *TeamStrategyService) ListTeamStrategyGroup(ctx context.Context, request *palace.ListTeamStrategyGroupRequest) (*palace.ListTeamStrategyGroupReply, error) {
	params := &bo.ListTeamStrategyGroupParams{
		Keyword:           request.GetKeyword(),
		Status:            vobj.GlobalStatus(request.GetStatus()),
		PaginationRequest: build.ToPaginationRequest(request.GetPagination()),
	}
	groups, err := t.teamStrategyGroupBiz.ListTeamStrategyGroup(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListTeamStrategyGroupReply{
		Items:      build.ToTeamStrategyGroupItems(groups.Items),
		Pagination: build.ToPaginationReply(groups.PaginationReply),
	}, nil
}

func (t *TeamStrategyService) SelectTeamStrategyGroup(ctx context.Context, request *palace.SelectTeamStrategyGroupRequest) (*palace.SelectTeamStrategyGroupReply, error) {
	params := build.ToSelectTeamStrategyGroupParams(request)
	groups, err := t.teamStrategyGroupBiz.SelectTeamStrategyGroup(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.SelectTeamStrategyGroupReply{
		Items:      build.ToSelectItems(groups.Items),
		Pagination: build.ToPaginationReply(groups.PaginationReply),
	}, nil
}

func (t *TeamStrategyService) SaveTeamStrategy(ctx context.Context, request *palace.SaveTeamStrategyRequest) (*common.EmptyReply, error) {
	params := build.ToSaveTeamStrategyParams(request)
	if err := t.teamStrategyBiz.SaveTeamStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) UpdateTeamStrategiesStatus(ctx context.Context, request *palace.UpdateTeamStrategiesStatusRequest) (*common.EmptyReply, error) {
	params := build.ToUpdateTeamStrategiesStatusParams(request)
	if err := t.teamStrategyBiz.UpdateTeamStrategiesStatus(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) DeleteTeamStrategy(ctx context.Context, request *palace.DeleteTeamStrategyRequest) (*common.EmptyReply, error) {
	if err := t.teamStrategyBiz.DeleteTeamStrategy(ctx, request.GetStrategyId()); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) ListTeamStrategy(ctx context.Context, request *palace.ListTeamStrategyRequest) (*palace.ListTeamStrategyReply, error) {
	params := build.ToListTeamStrategyParams(request)
	strategies, err := t.teamStrategyBiz.ListTeamStrategy(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.ListTeamStrategyReply{
		Items:      build.ToTeamStrategyItems(strategies.Items),
		Pagination: build.ToPaginationReply(strategies.PaginationReply),
	}, nil
}

func (t *TeamStrategyService) SubscribeTeamStrategy(ctx context.Context, request *palace.SubscribeTeamStrategyRequest) (*common.EmptyReply, error) {
	params := build.ToSubscribeTeamStrategyParams(request)
	if err := t.teamStrategyBiz.SubscribeTeamStrategy(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}

func (t *TeamStrategyService) SubscribeTeamStrategies(ctx context.Context, request *palace.SubscribeTeamStrategiesRequest) (*palace.SubscribeTeamStrategiesReply, error) {
	params := build.ToSubscribeTeamStrategiesParams(request)
	subscribers, err := t.teamStrategyBiz.SubscribeTeamStrategies(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.SubscribeTeamStrategiesReply{
		Items:      build.ToSubscribeTeamStrategiesItems(subscribers.Items),
		Pagination: build.ToPaginationReply(subscribers.PaginationReply),
	}, nil
}
