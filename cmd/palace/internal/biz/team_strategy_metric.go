package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
)

func NewTeamStrategyMetricBiz(
	teamStrategyRepo repository.TeamStrategy,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel,
	dictRepo repository.TeamDict,
	noticeGroupRepo repository.TeamNotice,
	datasourceRepo repository.TeamDatasourceMetric,
	transaction repository.Transaction,
) *TeamStrategyMetric {
	return &TeamStrategyMetric{
		teamStrategyRepo:            teamStrategyRepo,
		teamStrategyMetricRepo:      teamStrategyMetricRepo,
		teamStrategyMetricLevelRepo: teamStrategyMetricLevelRepo,
		dictRepo:                    dictRepo,
		noticeGroupRepo:             noticeGroupRepo,
		datasourceRepo:              datasourceRepo,
		transaction:                 transaction,
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
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategy(ctx context.Context, params *bo.SaveTeamMetricStrategyParams) error {
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
		if params.StrategyMetricID <= 0 {
			if err := params.Validate(); err != nil {
				return err
			}
			return t.teamStrategyMetricRepo.Create(ctx, params)
		}
		strategyMetricDo, err := t.teamStrategyMetricRepo.Get(ctx, params.StrategyID)
		if err != nil {
			return err
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
	return t.teamStrategyMetricLevelRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyMetric) DeleteTeamMetricStrategyLevel(ctx context.Context, strategyMetricLevelID uint32) error {
	return t.teamStrategyMetricLevelRepo.Delete(ctx, strategyMetricLevelID)
}

func (t *TeamStrategyMetric) GetTeamMetricStrategyLevel(ctx context.Context, strategyMetricLevelID uint32) (do.StrategyMetricRule, error) {
	return t.teamStrategyMetricLevelRepo.Get(ctx, strategyMetricLevelID)
}
