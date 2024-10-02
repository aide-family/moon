package repoimpl

import (
	"context"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel/alarmquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// NewAlarmHistoryRepository 创建告警记录仓库
func NewAlarmHistoryRepository(data *data.Data) repository.HistoryRepository {
	return &alarmHistoryRepositoryImpl{
		data: data,
	}
}

type (
	alarmHistoryRepositoryImpl struct {
		data *data.Data
	}
)

func (a *alarmHistoryRepositoryImpl) CreateAlarmHistory(ctx context.Context, param *bo.CreateAlarmInfoParams) error {
	teamID := param.TeamID
	alarmQuery, err := getTeamBizAlarmQuery(teamID, a.data)
	if err != nil {
		return err
	}
	historyList, err := a.createAlarmHistoryToModels(ctx, param)
	if err != nil {
		return err
	}

	if err := alarmQuery.AlarmHistory.WithContext(ctx).CreateInBatches(historyList, 100); err != nil {
		return err
	}

	return nil
}

func (a *alarmHistoryRepositoryImpl) GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	alarmHistory, err := alarmQuery.AlarmHistory.WithContext(ctx).Preload(field.Associations).Where(alarmQuery.AlarmHistory.ID.Eq(param.ID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nAlarmHistoryDataNotFoundErr(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return alarmHistory, nil
}

func (a *alarmHistoryRepositoryImpl) GetAlarmHistories(ctx context.Context, param *bo.QueryAlarmHistoryListParams) ([]*alarmmodel.AlarmHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	bizWrapper := alarmQuery.AlarmHistory.WithContext(ctx)
	var wheres []gen.Condition

	if !param.AlertStatus.IsUnknown() {
		wheres = append(wheres, alarmQuery.AlarmHistory.AlertStatus.Like(param.AlertStatus.GetValue()))
	}

	if !types.TextIsNull(param.Keyword) {
		bizWrapper = bizWrapper.Or(alarmQuery.AlarmHistory.Expr.Like(param.Keyword))
	}

	bizWrapper = bizWrapper.Where(wheres...)
	if err := types.WithPageQuery[alarmquery.IAlarmHistoryDo](bizWrapper, param.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(alarmQuery.AlarmHistory.ID.Desc()).Find()
}

func (a *alarmHistoryRepositoryImpl) createAlarmHistoryToModels(ctx context.Context, param *bo.CreateAlarmInfoParams) ([]*alarmmodel.AlarmHistory, error) {
	strategyID := param.StrategyID
	levelID := param.LevelID
	teamID := param.TeamID
	bizQuery, err := getTeamIdBizQuery(a.data, teamID)
	if err != nil {
		return nil, err
	}
	if !types.IsNil(err) {
		return nil, err
	}
	// 获取告警策略
	strategy, err := bizQuery.Strategy.WithContext(ctx).Preload(field.Associations).Where(bizQuery.Strategy.ID.Eq(strategyID)).First()
	if !types.IsNil(err) {
		return nil, err
	}
	// 获取level
	strategyLevel, err := bizQuery.StrategyLevel.WithContext(ctx).Preload(field.Associations).Where(bizQuery.StrategyLevel.ID.Eq(levelID)).First()
	if !types.IsNil(err) {
		return nil, err
	}

	historyList := types.SliceTo(param.Alerts, func(alarmParam *bo.CreateAlarmItemParams) *alarmmodel.AlarmHistory {
		return &alarmmodel.AlarmHistory{
			AlertStatus: vobj.ToAlertStatus(alarmParam.Status),
			StartsAt:    alarmParam.StartsAt,
			EndsAt:      alarmParam.EndsAt,
			Expr:        strategy.Expr,
			Fingerprint: alarmParam.Fingerprint,
			Labels:      vobj.NewLabels(alarmParam.Labels),
			Annotations: alarmParam.Annotations,
			RawInfoID:   param.RawInfoID,
			HistoryDetails: &alarmmodel.HistoryDetails{
				Strategy:   strategy.String(),
				Level:      strategyLevel.String(),
				Datasource: "",
			},
		}
	})
	return historyList, nil
}
