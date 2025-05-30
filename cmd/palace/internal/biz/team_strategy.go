package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

func NewTeamStrategyBiz(
	teamStrategyGroupRepo repository.TeamStrategyGroup,
	teamStrategyRepo repository.TeamStrategy,
	teamNoticeRepo repository.TeamNotice,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel,
	transactionRepo repository.Transaction,
) *TeamStrategy {
	return &TeamStrategy{
		teamStrategyGroupRepo:       teamStrategyGroupRepo,
		teamStrategyRepo:            teamStrategyRepo,
		teamNoticeRepo:              teamNoticeRepo,
		teamStrategyMetricRepo:      teamStrategyMetricRepo,
		teamStrategyMetricLevelRepo: teamStrategyMetricLevelRepo,
		transactionRepo:             transactionRepo,
	}
}

type TeamStrategy struct {
	teamStrategyGroupRepo       repository.TeamStrategyGroup
	teamStrategyRepo            repository.TeamStrategy
	teamStrategyMetricRepo      repository.TeamStrategyMetric
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel
	teamNoticeRepo              repository.TeamNotice
	transactionRepo             repository.Transaction
}

func (t *TeamStrategy) SaveTeamStrategy(ctx context.Context, params *bo.SaveTeamStrategyParams) error {
	if err := t.teamStrategyRepo.NameExists(ctx, params.Name, params.ID); err != nil {
		return err
	}
	strategyGroup, err := t.teamStrategyGroupRepo.Get(ctx, params.StrategyGroupID)
	if err != nil {
		return err
	}
	receiverRoutes, err := t.teamNoticeRepo.FindByIds(ctx, params.ReceiverRoutes)
	if err != nil {
		return err
	}

	params.WithStrategyGroup(strategyGroup)
	params.WithReceiverRoutes(receiverRoutes)

	return t.transactionRepo.BizExec(ctx, func(ctx context.Context) error {
		if params.ID <= 0 {
			if err := params.Validate(); err != nil {
				return err
			}
			return t.teamStrategyRepo.Create(ctx, params)
		}
		strategyDo, err := t.teamStrategyRepo.Get(ctx, params.ID)
		if err != nil {
			return err
		}

		params.WithStrategy(strategyDo)
		if err := params.Validate(); err != nil {
			return err
		}
		return t.teamStrategyRepo.Update(ctx, params)
	})
}

func (t *TeamStrategy) DeleteTeamStrategy(ctx context.Context, strategyId uint32) error {
	return t.transactionRepo.BizExec(ctx, func(ctx context.Context) error {
		if err := t.teamStrategyRepo.Delete(ctx, strategyId); err != nil {
			return err
		}
		if err := t.teamStrategyMetricRepo.DeleteByStrategyIds(ctx, strategyId); err != nil {
			return err
		}
		if err := t.teamStrategyMetricLevelRepo.DeleteByStrategyIds(ctx, strategyId); err != nil {
			return err
		}
		return nil
	})
}

func (t *TeamStrategy) UpdateTeamStrategiesStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error {
	return t.teamStrategyRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategy) ListTeamStrategy(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error) {
	return t.teamStrategyRepo.List(ctx, params)
}

func (t *TeamStrategy) SubscribeTeamStrategy(ctx context.Context, params *bo.SubscribeTeamStrategyParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return t.teamStrategyRepo.Subscribe(ctx, params)
}

func (t *TeamStrategy) SubscribeTeamStrategies(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error) {
	return t.teamStrategyRepo.SubscribeList(ctx, params)
}
