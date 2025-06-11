package biz

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

func NewTeamStrategyMetric(
	teamStrategyGroupRepo repository.TeamStrategyGroup,
	teamStrategyRepo repository.TeamStrategy,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	dictRepo repository.TeamDict,
	noticeGroupRepo repository.TeamNotice,
	datasourceRepo repository.TeamDatasourceMetric,
	transaction repository.Transaction,
) *TeamStrategyMetric {
	return &TeamStrategyMetric{
		teamStrategyGroupRepo:  teamStrategyGroupRepo,
		teamStrategyRepo:       teamStrategyRepo,
		teamStrategyMetricRepo: teamStrategyMetricRepo,
		dictRepo:               dictRepo,
		noticeGroupRepo:        noticeGroupRepo,
		datasourceRepo:         datasourceRepo,
		transaction:            transaction,
	}
}

type TeamStrategyMetric struct {
	teamStrategyGroupRepo  repository.TeamStrategyGroup
	teamStrategyRepo       repository.TeamStrategy
	teamStrategyMetricRepo repository.TeamStrategyMetric
	dictRepo               repository.TeamDict
	noticeGroupRepo        repository.TeamNotice
	datasourceRepo         repository.TeamDatasourceMetric
	transaction            repository.Transaction
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategy(ctx context.Context, params *bo.SaveTeamMetricStrategyParams) (do.StrategyMetric, error) {
	strategyDo, err := t.teamStrategyRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
	if err != nil {
		return nil, err
	}
	if strategyDo.GetStatus().IsEnable() {
		return nil, merr.ErrorBadRequest("策略已启用，不能修改")
	}
	datasourceDos, err := t.datasourceRepo.FindByIds(ctx, params.Datasource)
	if err != nil {
		return nil, err
	}
	var strategyMetricDo do.StrategyMetric
	err = t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if params.ID <= 0 {
			req := params.ToCreateTeamMetricStrategyParams(strategyDo, datasourceDos)
			if err := req.Validate(); err != nil {
				return err
			}
			strategyMetricDo, err = t.teamStrategyMetricRepo.Create(ctx, req)
			return err
		}
		strategyMetricDo, err = t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
		if err != nil {
			return err
		}
		req := params.ToUpdateTeamMetricStrategyParams(strategyDo, datasourceDos, strategyMetricDo)
		if err := req.Validate(); err != nil {
			return err
		}
		strategyMetricDo, err = t.teamStrategyMetricRepo.Update(ctx, req)
		return err
	})
	if err != nil {
		return nil, err
	}
	return strategyMetricDo, nil
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategyLevels(ctx context.Context, params *bo.SaveTeamMetricStrategyLevelsParams) ([]do.StrategyMetricRule, error) {
	strategyMetricDo, err := t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyMetricID})
	if err != nil {
		return nil, err
	}
	if strategyMetricDo.GetStrategy().GetStatus().IsEnable() {
		return nil, merr.ErrorBadRequest("策略已启用，不能修改")
	}
	noticeGroupIds := make([]uint32, 0, len(params.Levels))
	dictIds := make([]uint32, 0, len(params.Levels))
	for _, rule := range params.Levels {
		noticeGroupIds = append(noticeGroupIds, rule.GetNoticeGroupIds()...)
		dictIds = append(dictIds, rule.GetDictIds()...)
	}
	noticeGroupDos, err := t.noticeGroupRepo.FindByIds(ctx, noticeGroupIds)
	if err != nil {
		return nil, err
	}

	dicts, err := t.dictRepo.FindByIds(ctx, dictIds)
	if err != nil {
		return nil, err
	}
	params.ToSaveTeamMetricStrategyLevelsParams(strategyMetricDo, noticeGroupDos, dicts)
	if err := params.Validate(); err != nil {
		return nil, err
	}
	updatedRulesParams := &bo.SaveTeamMetricStrategyLevelsParams{
		StrategyMetricID: params.StrategyMetricID,
		Levels:           make([]*bo.SaveTeamMetricStrategyLevelParams, 0, len(params.Levels)),
	}
	createdRulesParams := &bo.SaveTeamMetricStrategyLevelsParams{
		StrategyMetricID: params.StrategyMetricID,
		Levels:           make([]*bo.SaveTeamMetricStrategyLevelParams, 0, len(params.Levels)),
	}
	for _, rule := range params.Levels {
		if rule.GetID() <= 0 {
			createdRulesParams.Levels = append(createdRulesParams.Levels, rule)
		} else {
			updatedRulesParams.Levels = append(updatedRulesParams.Levels, rule)
		}
	}
	updatedRulesParams.ToSaveTeamMetricStrategyLevelsParams(strategyMetricDo, noticeGroupDos, dicts)
	createdRulesParams.ToSaveTeamMetricStrategyLevelsParams(strategyMetricDo, noticeGroupDos, dicts)
	if err := updatedRulesParams.Validate(); err != nil {
		return nil, err
	}
	if err := createdRulesParams.Validate(); err != nil {
		return nil, err
	}
	levels, err := t.teamStrategyMetricRepo.FindLevels(ctx, &bo.FindTeamMetricStrategyLevelsParams{
		StrategyMetricID: params.StrategyMetricID,
	})
	if err != nil {
		return nil, err
	}
	levelIds := slices.Map(levels, func(v do.StrategyMetricRule) uint32 { return v.GetID() })
	list := make([]do.StrategyMetricRule, 0, len(updatedRulesParams.Levels))
	err = t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if len(updatedRulesParams.Levels) > 0 {
			updatedRules, err := t.teamStrategyMetricRepo.UpdateLevels(ctx, updatedRulesParams)
			if err != nil {
				return err
			}
			list = append(list, updatedRules...)
		}
		if len(createdRulesParams.Levels) > 0 {
			createdRules, err := t.teamStrategyMetricRepo.CreateLevels(ctx, createdRulesParams)
			if err != nil {
				return err
			}
			list = append(list, createdRules...)
		}
		deleteIds := make([]uint32, 0, len(levelIds))
		existingMap := slices.ToMap(list, func(v do.StrategyMetricRule) uint32 { return v.GetID() })
		for _, ruleId := range levelIds {
			if _, ok := existingMap[ruleId]; !ok {
				deleteIds = append(deleteIds, ruleId)
			}
		}
		return t.teamStrategyMetricRepo.DeleteUnUsedLevels(ctx, &bo.DeleteUnUsedLevelsParams{
			StrategyMetricID: params.StrategyMetricID,
			RuleIds:          deleteIds,
		})
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (t *TeamStrategyMetric) GetTeamMetricStrategy(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.StrategyMetric, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return t.teamStrategyMetricRepo.Get(ctx, params)
}
