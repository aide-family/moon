package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewTeamStrategyMetricBiz(
	teamStrategyRepo repository.TeamStrategy,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel,
	dictRepo repository.TeamDict,
	noticeGroupRepo repository.TeamNotice,
	datasourceRepo repository.TeamDatasourceMetric,
	transaction repository.Transaction,
	eventBus repository.EventBus,
	logger log.Logger,
) *TeamStrategyMetric {
	return &TeamStrategyMetric{
		teamStrategyRepo:            teamStrategyRepo,
		teamStrategyMetricRepo:      teamStrategyMetricRepo,
		teamStrategyMetricLevelRepo: teamStrategyMetricLevelRepo,
		dictRepo:                    dictRepo,
		noticeGroupRepo:             noticeGroupRepo,
		datasourceRepo:              datasourceRepo,
		transaction:                 transaction,
		eventBus:                    eventBus,
		helper:                      log.NewHelper(log.With(logger, "module", "biz.team_strategy_metric")),
	}
}

type TeamStrategyMetric struct {
	teamStrategyRepo            repository.TeamStrategy
	teamStrategyMetricRepo      repository.TeamStrategyMetric
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel
	dictRepo                    repository.TeamDict
	noticeGroupRepo             repository.TeamNotice
	datasourceRepo              repository.TeamDatasourceMetric
	transaction                 repository.Transaction
	eventBus                    repository.EventBus
	helper                      *log.Helper
}

func (t *TeamStrategyMetric) publishStrategyDataChangeEvent(ctx context.Context, strategyIds ...uint32) {
	if len(strategyIds) == 0 {
		return
	}
	strategyIds = slices.Unique(strategyIds)
	teamID := permission.GetTeamIDByContextWithZeroValue(safety.CopyValueCtx(ctx))
	go func(teamID uint32, ids ...uint32) {
		defer func() {
			if r := recover(); r != nil {
				t.helper.Errorw("publishDataChangeEvent", "error", r)
			}
		}()
		t.eventBus.PublishDataChangeEvent(vobj.ChangedTypeMetricStrategy, teamID, ids...)
	}(teamID, strategyIds...)
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategy(ctx context.Context, params *bo.SaveTeamMetricStrategyParams) error {
	defer t.publishStrategyDataChangeEvent(ctx, params.StrategyID)
	strategyDo, err := t.teamStrategyRepo.Get(ctx, params.StrategyID)
	if err != nil {
		return err
	}
	params.WithStrategy(strategyDo)
	datasourceDos, err := t.datasourceRepo.FindByIds(ctx, params.Datasource)
	if err != nil {
		return err
	}
	params.WithDatasource(datasourceDos)
	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		strategyMetricDo, err := t.teamStrategyMetricRepo.GetByStrategyId(ctx, params.StrategyID)
		if err != nil {
			if !merr.IsNotFound(err) {
				return err
			}
			if err := params.Validate(); err != nil {
				return err
			}
			return t.teamStrategyMetricRepo.Create(ctx, params)
		}

		params.WithStrategyMetric(strategyMetricDo)
		if err := params.Validate(); err != nil {
			return err
		}
		return t.teamStrategyMetricRepo.Update(ctx, params)
	})
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategyLevel(ctx context.Context, params *bo.SaveTeamMetricStrategyLevelParams) error {
	strategyMetricDo, err := t.teamStrategyMetricRepo.Get(ctx, params.StrategyMetricID)
	if err != nil {
		return err
	}
	defer t.publishStrategyDataChangeEvent(ctx, strategyMetricDo.GetStrategyID())
	if strategyMetricDo.GetStrategy().GetStatus().IsEnable() {
		return merr.ErrorBadRequest("strategy is enabled and cannot be modified")
	}

	noticeGroupDos, err := t.noticeGroupRepo.FindByIds(ctx, params.GetNoticeGroupIds())
	if err != nil {
		return err
	}

	dictDos, err := t.dictRepo.FindByIds(ctx, params.GetDictIds())
	if err != nil {
		return err
	}
	params.WithStrategyMetric(strategyMetricDo)
	params.WithNoticeGroupDos(noticeGroupDos)
	params.WithDicts(dictDos)

	return t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if params.StrategyMetricLevelID <= 0 {
			if err := params.Validate(); err != nil {
				return err
			}
			return t.teamStrategyMetricLevelRepo.Create(ctx, params)
		}
		strategyMetricLevelDo, err := t.teamStrategyMetricLevelRepo.Get(ctx, params.StrategyMetricLevelID)
		if err != nil {
			return err
		}
		params.WithStrategyMetricLevel(strategyMetricLevelDo)
		if err := params.Validate(); err != nil {
			return err
		}
		return t.teamStrategyMetricLevelRepo.Update(ctx, params)
	})
}

func (t *TeamStrategyMetric) GetTeamMetricStrategy(ctx context.Context, strategyId uint32) (do.StrategyMetric, error) {
	return t.teamStrategyMetricRepo.Get(ctx, strategyId)
}

func (t *TeamStrategyMetric) GetTeamMetricStrategyByStrategyId(ctx context.Context, strategyId uint32) (do.StrategyMetric, error) {
	return t.teamStrategyMetricRepo.GetByStrategyId(ctx, strategyId)
}

func (t *TeamStrategyMetric) ListTeamMetricStrategyLevels(ctx context.Context, params *bo.ListTeamMetricStrategyLevelsParams) (*bo.ListTeamMetricStrategyLevelsReply, error) {
	return t.teamStrategyMetricLevelRepo.List(ctx, params)
}

func (t *TeamStrategyMetric) UpdateTeamMetricStrategyLevelStatus(ctx context.Context, params *bo.UpdateTeamMetricStrategyLevelStatusParams) error {
	strategyMetricLevelDos, err := t.teamStrategyMetricLevelRepo.FindByIds(ctx, params.StrategyMetricLevelIds)
	if err != nil {
		return err
	}
	strategyIds := slices.Map(strategyMetricLevelDos, func(item do.StrategyMetricRule) uint32 {
		return item.GetStrategyID()
	})
	defer t.publishStrategyDataChangeEvent(ctx, strategyIds...)
	return t.teamStrategyMetricLevelRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyMetric) DeleteTeamMetricStrategyLevels(ctx context.Context, strategyMetricLevelIds []uint32) error {
	strategyMetricLevelDos, err := t.teamStrategyMetricLevelRepo.FindByIds(ctx, strategyMetricLevelIds)
	if err != nil {
		return err
	}
	strategyIds := slices.Map(strategyMetricLevelDos, func(item do.StrategyMetricRule) uint32 {
		return item.GetStrategyID()
	})
	defer t.publishStrategyDataChangeEvent(ctx, strategyIds...)
	return t.teamStrategyMetricLevelRepo.Delete(ctx, strategyMetricLevelIds)
}

func (t *TeamStrategyMetric) GetTeamMetricStrategyLevel(ctx context.Context, strategyMetricLevelID uint32) (do.StrategyMetricRule, error) {
	return t.teamStrategyMetricLevelRepo.Get(ctx, strategyMetricLevelID)
}
