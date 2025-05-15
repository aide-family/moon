package biz

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
)

func NewTeamStrategy(
	teamStrategyGroupRepo repository.TeamStrategyGroup,
	teamStrategyRepo repository.TeamStrategy,
	teamNoticeRepo repository.TeamNotice,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	transactionRepo repository.Transaction,
) *TeamStrategy {
	return &TeamStrategy{
		teamStrategyGroupRepo:  teamStrategyGroupRepo,
		teamStrategyRepo:       teamStrategyRepo,
		teamNoticeRepo:         teamNoticeRepo,
		teamStrategyMetricRepo: teamStrategyMetricRepo,
		transactionRepo:        transactionRepo,
	}
}

type TeamStrategy struct {
	teamStrategyGroupRepo  repository.TeamStrategyGroup
	teamStrategyRepo       repository.TeamStrategy
	teamStrategyMetricRepo repository.TeamStrategyMetric
	teamNoticeRepo         repository.TeamNotice
	transactionRepo        repository.Transaction
}

func (t *TeamStrategy) SaveTeamStrategy(ctx context.Context, params *bo.SaveTeamStrategyParams) (do.Strategy, error) {
	strategyGroup, err := t.teamStrategyGroupRepo.Get(ctx, params.StrategyGroupID)
	if err != nil {
		return nil, err
	}
	receiverRoutes, err := t.teamNoticeRepo.FindByIds(ctx, params.ReceiverRoutes)
	if err != nil {
		return nil, err
	}

	var strategyDo do.Strategy
	err = t.transactionRepo.BizExec(ctx, func(ctx context.Context) error {
		if params.ID <= 0 {
			req := params.ToCreateTeamStrategyParams(strategyGroup, receiverRoutes)
			if err := req.Validate(); err != nil {
				return err
			}
			strategyDo, err = t.teamStrategyRepo.Create(ctx, req)
			return err
		}
		strategyDo, err = t.teamStrategyRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.ID})
		if err != nil {
			return err
		}

		req := params.ToUpdateTeamStrategyParams(strategyGroup, strategyDo, receiverRoutes)
		if err := req.Validate(); err != nil {
			return err
		}
		strategyDo, err = t.teamStrategyRepo.Update(ctx, req)
		return err
	})
	if err != nil {
		return nil, err
	}
	return strategyDo, nil
}

func (t *TeamStrategy) DeleteTeamStrategy(ctx context.Context, params *bo.OperateTeamStrategyParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return t.transactionRepo.BizExec(ctx, func(ctx context.Context) error {
		if err := t.teamStrategyRepo.Delete(ctx, params); err != nil {
			return err
		}
		if err := t.teamStrategyMetricRepo.Delete(ctx, params); err != nil {
			return err
		}
		if err := t.teamStrategyMetricRepo.DeleteLevels(ctx, params); err != nil {
			return err
		}
		return nil
	})
}

func (t *TeamStrategy) UpdateTeamStrategiesStatus(ctx context.Context, params *bo.UpdateTeamStrategiesStatusParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return t.teamStrategyRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategy) ListTeamStrategy(ctx context.Context, params *bo.ListTeamStrategyParams) (*bo.ListTeamStrategyReply, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return t.teamStrategyRepo.List(ctx, params)
}

func (t *TeamStrategy) SubscribeTeamStrategy(ctx context.Context, params *bo.SubscribeTeamStrategyParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return t.teamStrategyRepo.Subscribe(ctx, params)
}

func (t *TeamStrategy) SubscribeTeamStrategies(ctx context.Context, params *bo.SubscribeTeamStrategiesParams) (*bo.SubscribeTeamStrategiesReply, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return t.teamStrategyRepo.SubscribeList(ctx, params)
}
