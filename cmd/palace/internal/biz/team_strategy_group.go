package biz

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
)

func NewTeamStrategyGroupBiz(teamStrategyGroupRepo repository.TeamStrategyGroup) *TeamStrategyGroupBiz {
	return &TeamStrategyGroupBiz{
		teamStrategyGroupRepo: teamStrategyGroupRepo,
	}
}

type TeamStrategyGroupBiz struct {
	teamStrategyGroupRepo repository.TeamStrategyGroup
}

func (t *TeamStrategyGroupBiz) SaveTeamStrategyGroup(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	if params.ID <= 0 {
		return t.teamStrategyGroupRepo.Create(ctx, params)
	}
	return t.teamStrategyGroupRepo.Update(ctx, params)
}

func (t *TeamStrategyGroupBiz) UpdateTeamStrategyGroupStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error {
	return t.teamStrategyGroupRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyGroupBiz) DeleteTeamStrategyGroup(ctx context.Context, id uint32) error {
	return t.teamStrategyGroupRepo.Delete(ctx, id)
}

func (t *TeamStrategyGroupBiz) GetTeamStrategyGroup(ctx context.Context, id uint32) (do.StrategyGroup, error) {
	return t.teamStrategyGroupRepo.Get(ctx, id)
}

func (t *TeamStrategyGroupBiz) ListTeamStrategyGroup(ctx context.Context, params *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error) {
	return t.teamStrategyGroupRepo.List(ctx, params)
}
