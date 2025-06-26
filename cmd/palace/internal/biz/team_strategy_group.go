package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewTeamStrategyGroupBiz(
	teamStrategyGroupRepo repository.TeamStrategyGroup,
	teamStrategyRepo repository.TeamStrategy,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel,
	transaction repository.Transaction,
	eventBus repository.EventBus,
	logger log.Logger,
) *TeamStrategyGroup {
	return &TeamStrategyGroup{
		teamStrategyGroupRepo:       teamStrategyGroupRepo,
		teamStrategyRepo:            teamStrategyRepo,
		teamStrategyMetricRepo:      teamStrategyMetricRepo,
		teamStrategyMetricLevelRepo: teamStrategyMetricLevelRepo,
		transaction:                 transaction,
		eventBus:                    eventBus,
		helper:                      log.NewHelper(log.With(logger, "module", "biz.team_strategy_group")),
	}
}

type TeamStrategyGroup struct {
	teamStrategyGroupRepo       repository.TeamStrategyGroup
	teamStrategyRepo            repository.TeamStrategy
	teamStrategyMetricRepo      repository.TeamStrategyMetric
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel
	transaction                 repository.Transaction
	eventBus                    repository.EventBus
	helper                      *log.Helper
}

func (t *TeamStrategyGroup) publishStrategyDataChangeEvent(ctx context.Context, groupIds ...uint32) {
	if len(groupIds) == 0 {
		return
	}
	groupIds = slices.Unique(groupIds)
	ctx = safety.CopyValueCtx(ctx)
	teamID := permission.GetTeamIDByContextWithZeroValue(ctx)
	safety.Go("publishStrategyDataChangeEvent", func() {
		// query strategy ids by group ids
		strategyDos, err := t.teamStrategyRepo.FindByStrategiesGroupIds(ctx, groupIds...)
		if err != nil {
			t.helper.Errorw("publishStrategyDataChangeEvent", "error", err)
			return
		}
		strategyIds := slices.Map(strategyDos, func(item do.Strategy) uint32 {
			return item.GetID()
		})
		strategyIds = slices.Unique(strategyIds)
		if len(strategyIds) > 0 {
			t.eventBus.PublishDataChangeEvent(vobj.ChangedTypeMetricStrategy, teamID, strategyIds...)
		}
	}, t.helper.Logger())
}

func (t *TeamStrategyGroup) SaveTeamStrategyGroup(ctx context.Context, params *bo.SaveTeamStrategyGroupParams) error {
	if params.ID <= 0 {
		return t.teamStrategyGroupRepo.Create(ctx, params)
	}
	return t.teamStrategyGroupRepo.Update(ctx, params)
}

func (t *TeamStrategyGroup) UpdateTeamStrategyGroupStatus(ctx context.Context, params *bo.UpdateTeamStrategyGroupStatusParams) error {
	defer t.publishStrategyDataChangeEvent(ctx, params.ID)
	return t.teamStrategyGroupRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyGroup) DeleteTeamStrategyGroup(ctx context.Context, id uint32) error {
	defer t.publishStrategyDataChangeEvent(ctx, id)
	strategyDo, err := t.teamStrategyRepo.FindByStrategiesGroupIds(ctx, id)
	if err != nil {
		return err
	}
	strategyIds := slices.Map(strategyDo, func(item do.Strategy) uint32 {
		return item.GetID()
	})
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if err := t.teamStrategyRepo.DeleteByStrategyIds(ctx, strategyIds...); err != nil {
			return err
		}
		if err := t.teamStrategyMetricRepo.DeleteByStrategyIds(ctx, strategyIds...); err != nil {
			return err
		}
		if err := t.teamStrategyMetricLevelRepo.DeleteByStrategyIds(ctx, strategyIds...); err != nil {
			return err
		}
		return t.teamStrategyGroupRepo.Delete(ctx, id)
	})
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
