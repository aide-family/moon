package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewRealtimeAlarmRepository 实例化告警业务数据库
func NewRealtimeAlarmRepository(data *data.Data) repository.Alarm {
	return &realtimeAlarmRepositoryImpl{data: data}
}

type realtimeAlarmRepositoryImpl struct {
	data *data.Data
}

func (r *realtimeAlarmRepositoryImpl) CreateRealTimeAlarm(ctx context.Context, param *bo.CreateAlarmInfoParams) error {
	alarmQuery, err := getTeamBizAlarmQuery(param.TeamID, r.data)
	if err != nil {
		return err
	}

	realTimes := r.createRealTimeAlarmToModels(param)
	// 更新的字段
	columns := []string{"summary", "description", "status", "starts_at", "ends_at", "expr", "labels", "annotations"}
	if err := alarmQuery.RealtimeAlarm.WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "fingerprint"}},
			DoUpdates: clause.AssignmentColumns(columns)}).
		CreateInBatches(realTimes, len(realTimes)); err != nil {
		return err
	}
	return nil
}

func (r *realtimeAlarmRepositoryImpl) SaveAlertQueue(param *bo.CreateAlarmHookRawParams) error {
	queue := r.data.GetAlertPersistenceDBQueue()
	if err := queue.Push(param.Message()); err != nil {
		return err
	}
	return nil
}

func (r *realtimeAlarmRepositoryImpl) GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*alarmmodel.RealtimeAlarm, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, r.data)
	if err != nil {
		return nil, err
	}
	var wheres []gen.Condition
	if !types.TextIsNull(params.Fingerprint) {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.Fingerprint.Eq(params.Fingerprint))
	}
	if params.RealtimeAlarmID != 0 {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.ID.Eq(params.RealtimeAlarmID))
	}
	detail, err := alarmQuery.WithContext(ctx).RealtimeAlarm.Preload(field.Associations).Where(wheres...).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastRealtimeAlarmNotFound(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return detail, nil
}

func (r *realtimeAlarmRepositoryImpl) GetRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*alarmmodel.RealtimeAlarm, error) {
	alarmQuery, err := getBizAlarmQuery(ctx, r.data)
	if err != nil {
		return nil, err
	}
	var wheres []gen.Condition
	if !types.TextIsNull(params.EventAtStart) && !types.TextIsNull(params.EventAtEnd) {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.StartsAt.Between(params.EventAtStart, params.EventAtEnd))
	}
	if !types.TextIsNull(params.ResolvedAtStart) && !types.TextIsNull(params.ResolvedAtEnd) {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.Status.Eq(vobj.AlertStatusResolved.GetValue()))
		wheres = append(wheres, alarmQuery.RealtimeAlarm.EndsAt.Between(params.ResolvedAtStart, params.ResolvedAtEnd))
	}

	if len(params.AlarmStatuses) > 0 {
		statuses := types.SliceTo(params.AlarmStatuses, func(status vobj.AlertStatus) int {
			return status.GetValue()
		})
		wheres = append(wheres, alarmQuery.RealtimeAlarm.Status.In(statuses...))
	}
	// TODO 获取指定告警页面告警数据
	// TODO 获取指定人员告警数据
	realtimeAlarmQuery := alarmQuery.WithContext(ctx).RealtimeAlarm.Where(wheres...)
	if realtimeAlarmQuery, err = types.WithPageQuery(realtimeAlarmQuery, params.Pagination); err != nil {
		return nil, err
	}
	return realtimeAlarmQuery.Order(alarmQuery.RealtimeAlarm.ID.Desc()).Find()
}

func (r *realtimeAlarmRepositoryImpl) createRealTimeAlarmToModels(param *bo.CreateAlarmInfoParams) []*alarmmodel.RealtimeAlarm {
	strategy := param.Strategy
	strategyLevel := param.Level

	alarms := types.SliceTo(param.Alerts, func(alarmParam *bo.AlertItemRawParams) *alarmmodel.RealtimeAlarm {
		labels := vobj.NewLabels(alarmParam.Labels)
		annotations := alarmParam.Annotations
		return &alarmmodel.RealtimeAlarm{
			Summary:     annotations.GetSummary(),
			Description: annotations.GetDescription(),
			Status:      vobj.ToAlertStatus(alarmParam.Status),
			StartsAt:    alarmParam.StartsAt,
			EndsAt:      alarmParam.EndsAt,
			Expr:        strategy.Expr,
			Fingerprint: alarmParam.Fingerprint,
			Labels:      labels,
			Annotations: annotations,
			RawInfoID:   param.GetRawInfoId(alarmParam.Fingerprint),
			RealtimeDetails: &alarmmodel.RealtimeDetails{
				Strategy:   strategy.String(),
				Level:      strategyLevel.String(),
				Datasource: param.GetDatasourceMap(vobj.NewLabels(alarmParam.Labels).GetDatasourceID()),
			},
		}
	})
	return alarms
}
