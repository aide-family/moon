package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel/alarmquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	historyList := a.createAlarmHistoryToModels(param)
	// 所更新的字段
	historyCol := []string{"summary", "description", "status", "starts_at", "ends_at", "expr", "labels", "annotations"}
	detailCol := []string{"strategy", "level", "datasource"}
	return alarmQuery.Transaction(func(tx *alarmquery.Query) error {
		for _, history := range historyList {
			// 告警历史表
			if err := tx.AlarmHistory.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "fingerprint"}},
				DoUpdates: clause.AssignmentColumns(historyCol),
			}).Create(history); err != nil {
				return err
			}

			// 告警详情表
			detail := &alarmmodel.HistoryDetails{
				Strategy:       param.Strategy.String(),
				Level:          param.Level,
				Datasource:     param.GetDatasourceMap(history.Labels.GetDatasourceID()),
				AlarmHistoryID: history.ID,
				AlarmHistory:   history,
			}
			if err := tx.HistoryDetails.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "alarm_history_id"}},
				DoUpdates: clause.AssignmentColumns(detailCol),
			}).Create(detail); err != nil {
				return err
			}
		}
		return nil
	})
}

func (a *alarmHistoryRepositoryImpl) GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, a.data)
	if err != nil {
		return nil, err
	}
	alarmHistory, err := alarmQuery.AlarmHistory.WithContext(ctx).Preload(field.Associations).Where(alarmQuery.AlarmHistory.ID.Eq(param.ID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastHistoryAlarmNotFound(ctx).WithCause(err)
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

	if len(param.AlertStatus) > 0 {
		statuses := types.SliceTo(param.AlertStatus, func(status vobj.AlertStatus) int {
			return status.GetValue()
		})
		wheres = append(wheres, alarmQuery.AlarmHistory.AlertStatus.In(statuses...))
	}

	if !types.TextIsNull(param.Keyword) {
		bizWrapper = bizWrapper.Or(alarmQuery.AlarmHistory.Expr.Like(param.Keyword))
	}

	if !types.TextIsNull(param.EventAtStart) && !types.TextIsNull(param.EventAtEnd) {
		wheres = append(wheres, alarmQuery.AlarmHistory.StartsAt.Between(param.EventAtStart, param.EventAtEnd))
	}
	if !types.TextIsNull(param.ResolvedAtStart) && !types.TextIsNull(param.ResolvedAtEnd) {
		wheres = append(wheres, alarmQuery.AlarmHistory.AlertStatus.Eq(vobj.AlertStatusResolved.GetValue()))
		wheres = append(wheres, alarmQuery.AlarmHistory.EndsAt.Between(param.ResolvedAtStart, param.ResolvedAtEnd))
	}

	bizWrapper = bizWrapper.Where(wheres...)
	if bizWrapper, err = types.WithPageQuery(bizWrapper, param.Page); err != nil {
		return nil, err
	}
	return bizWrapper.Order(alarmQuery.AlarmHistory.ID.Desc()).Find()
}

func (a *alarmHistoryRepositoryImpl) createAlarmHistoryToModels(param *bo.CreateAlarmInfoParams) []*alarmmodel.AlarmHistory {
	strategy := param.Strategy
	historyList := types.SliceTo(param.Alerts, func(alarmParam *bo.AlertItemRawParams) *alarmmodel.AlarmHistory {
		labels := vobj.NewLabels(alarmParam.Labels)
		annotations := vobj.NewAnnotations(alarmParam.Annotations)
		alertStatus := vobj.ToAlertStatus(alarmParam.Status)
		return &alarmmodel.AlarmHistory{
			Summary:     annotations.GetSummary(),
			Description: annotations.GetDescription(),
			AlertStatus: alertStatus,
			StartsAt:    alarmParam.StartsAt,
			EndsAt:      alarmParam.EndsAt,
			Expr:        strategy.Expr,
			Fingerprint: alarmParam.Fingerprint,
			Labels:      labels,
			Annotations: annotations,
			RawInfoID:   param.GetRawInfoID(alarmParam.Fingerprint),
		}
	})
	return historyList
}
