package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

func NewTeamStrategyGroupBiz(teamStrategyGroupRepo repository.TeamStrategyGroup) *TeamStrategyGroup {
	return &TeamStrategyGroup{
		teamStrategyGroupRepo: teamStrategyGroupRepo,
	}
}

type TeamStrategyGroup struct {
	teamStrategyGroupRepo repository.TeamStrategyGroup
}

func (t *TeamStrategyGroup) SaveTeamStrategyGroup(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	if params.ID <= 0 {
		return t.teamStrategyGroupRepo.Create(ctx, params)
	}
	return t.teamStrategyGroupRepo.Update(ctx, params)
}

func (t *TeamStrategyGroup) UpdateTeamStrategyGroupStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error {
	return t.teamStrategyGroupRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyGroup) DeleteTeamStrategyGroup(ctx context.Context, id uint32) error {
	return t.teamStrategyGroupRepo.Delete(ctx, id)
}

func (t *TeamStrategyGroup) GetTeamStrategyGroup(ctx context.Context, id uint32) (do.StrategyGroup, error) {
	return t.teamStrategyGroupRepo.Get(ctx, id)
}

func (t *TeamStrategyGroup) ListTeamStrategyGroup(ctx context.Context, params *bo.ListTeamStrategyGroupParams) (*bo.ListTeamStrategyGroupReply, error) {
	return t.teamStrategyGroupRepo.List(ctx, params)
}

func (t *TeamStrategyGroup) SelectTeamStrategyGroup(ctx context.Context, params *bo.SelectTeamStrategyGroupRequest) (*bo.SelectTeamStrategyGroupReply, error) {
	return t.teamStrategyGroupRepo.Select(ctx, params)
}
