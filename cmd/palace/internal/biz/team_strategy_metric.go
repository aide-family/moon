package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewTeamStrategyMetricBiz(
	teamStrategyGroupRepo repository.TeamStrategyGroup,
	teamStrategyRepo repository.TeamStrategy,
	teamStrategyMetricRepo repository.TeamStrategyMetric,
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel,
	dictRepo repository.TeamDict,
	noticeGroupRepo repository.TeamNotice,
	datasourceRepo repository.TeamDatasourceMetric,
	transaction repository.Transaction,
) *TeamStrategyMetric {
	return &TeamStrategyMetric{
		teamStrategyGroupRepo:       teamStrategyGroupRepo,
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
	teamStrategyGroupRepo       repository.TeamStrategyGroup
	teamStrategyRepo            repository.TeamStrategy
	teamStrategyMetricRepo      repository.TeamStrategyMetric
	teamStrategyMetricLevelRepo repository.TeamStrategyMetricLevel
	dictRepo                    repository.TeamDict
	noticeGroupRepo             repository.TeamNotice
	datasourceRepo              repository.TeamDatasourceMetric
	transaction                 repository.Transaction
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategy(ctx context.Context, params *bo.SaveTeamMetricStrategyParams) (do.StrategyMetric, error) {
	strategyDo, err := t.teamStrategyRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
	if err != nil {
		return nil, err
	}
	if strategyDo.GetStatus().IsEnable() {
		return nil, merr.ErrorParams("strategy is enabled and cannot be modified")
	}
	datasourceDos, err := t.datasourceRepo.FindByIds(ctx, params.Datasource)
	if err != nil {
		return nil, err
	}
	if len(datasourceDos) == 0 {
		return nil, merr.ErrorParams("datasource not found")
	}
	strategyMetricDo, err := t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
	if err != nil && !merr.IsNotFound(err) {
		return nil, err
	}
	err = t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if validate.IsNil(strategyMetricDo) || strategyMetricDo.GetID() == 0 {
			req := params.ToCreateTeamMetricStrategyParams(strategyDo, datasourceDos)
			if err := req.Validate(); err != nil {
				return err
			}
			return t.teamStrategyMetricRepo.Create(ctx, req)
		}
		strategyMetricDo, err = t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
		if err != nil {
			return err
		}
		req := params.ToUpdateTeamMetricStrategyParams(strategyDo, datasourceDos, strategyMetricDo)
		if err := req.Validate(); err != nil {
			return err
		}
		return t.teamStrategyMetricRepo.Update(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	return t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyID})
}

func (t *TeamStrategyMetric) SaveTeamMetricStrategyLevel(ctx context.Context, params *bo.SaveTeamMetricStrategyLevelParams) (do.StrategyMetricRule, error) {
	strategyMetricDo, err := t.teamStrategyMetricRepo.Get(ctx, &bo.OperateTeamStrategyParams{StrategyId: params.StrategyMetricID})
	if err != nil {
		return nil, err
	}
	if strategyMetricDo.GetStrategy().GetStatus().IsEnable() {
		return nil, merr.ErrorBadRequest("strategy is enabled and cannot be modified")
	}
	noticeGroupIds := make([]uint32, 0, len(params.ReceiverRoutes)+len(params.LabelNotices))
	dictIds := make([]uint32, 0, len(params.AlarmPages))
	noticeGroupDos, err := t.noticeGroupRepo.FindByIds(ctx, noticeGroupIds)
	if err != nil {
		return nil, err
	}

	dictDos, err := t.dictRepo.FindByIds(ctx, dictIds)
	if err != nil {
		return nil, err
	}
	if err := params.Validate(); err != nil {
		return nil, err
	}
	saveParams := params.ToSaveTeamMetricStrategyLevelParams(strategyMetricDo, noticeGroupDos, dictDos)
	err = t.transaction.BizExec(ctx, func(ctx context.Context) error {
		if params.ID > 0 {
			return t.teamStrategyMetricLevelRepo.Update(ctx, saveParams)
		}
		return t.teamStrategyMetricLevelRepo.Create(ctx, saveParams)
	})
	if err != nil {
		return nil, err
	}
	detailParams := &bo.OperateTeamStrategyLevelParams{
		StrategyMetricId: params.StrategyMetricID,
		StrategyLevelId:  params.LevelId,
	}
	return t.teamStrategyMetricLevelRepo.GetByLevelId(ctx, detailParams)
}

func (t *TeamStrategyMetric) GetTeamMetricStrategy(ctx context.Context, params *bo.OperateTeamStrategyParams) (do.StrategyMetric, error) {
	return t.teamStrategyMetricRepo.Get(ctx, params)
}

func (t *TeamStrategyMetric) ListTeamMetricStrategyLevels(ctx context.Context, params *bo.ListTeamMetricStrategyLevelsParams) (*bo.ListTeamMetricStrategyLevelsReply, error) {
	return t.teamStrategyMetricLevelRepo.List(ctx, params)
}

func (t *TeamStrategyMetric) UpdateTeamMetricStrategyLevelStatus(ctx context.Context, params *bo.UpdateTeamMetricStrategyLevelStatusParams) error {
	return t.teamStrategyMetricLevelRepo.UpdateStatus(ctx, params)
}

func (t *TeamStrategyMetric) DeleteTeamMetricStrategyLevel(ctx context.Context, params *bo.DeleteTeamMetricStrategyLevelParams) error {
	return t.teamStrategyMetricLevelRepo.Delete(ctx, params.StrategyMetricLevelID)
}
