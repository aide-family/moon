package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel/alarmquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

// NewRealtimeAlarmRepository 实例化告警业务数据库
func NewRealtimeAlarmRepository(data *data.Data) repository.Alarm {
	return &realtimeAlarmRepositoryImpl{data: data}
}

type realtimeAlarmRepositoryImpl struct {
	data *data.Data
}

func (r *realtimeAlarmRepositoryImpl) CreateRealTimeAlarm(ctx context.Context, param *bo.CreateAlarmItemParams) error {
	alarmQuery, err := getBizAlarmQuery(ctx, r.data)
	if err != nil {
		return err
	}
	realTimeModel, err := r.createRealTimeAlarmToModel(ctx, param)
	if err != nil {
		return err
	}

	if err := alarmQuery.RealtimeAlarm.WithContext(ctx).Create(realTimeModel); err != nil {
		return err
	}
	return nil
}

// getBizQuery 获取告警业务数据库
func getBizAlarmQuery(ctx context.Context, data *data.Data) (*alarmquery.Query, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnauthorized(ctx)
	}
	bizDB, err := data.GetAlarmGormDB(claims.GetTeam())
	if !types.IsNil(err) {
		return nil, err
	}
	return alarmquery.Use(bizDB), nil
}
func (r *realtimeAlarmRepositoryImpl) SaveAlertQueue(param *bo.CreateAlarmHookRawParams) error {
	queue := r.data.GetAlartHistoryQueue()
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
	if params.EventAtStart < params.EventAtEnd && params.EventAtStart > 0 {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.StartsAt.Between(params.EventAtStart, params.EventAtEnd))
	}
	if params.ResolvedAtStart < params.ResolvedAtEnd && params.ResolvedAtStart > 0 {
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
	if err := types.WithPageQuery[alarmquery.IRealtimeAlarmDo](realtimeAlarmQuery, params.Pagination); err != nil {
		return nil, err
	}
	return realtimeAlarmQuery.Find()
}

func (r *realtimeAlarmRepositoryImpl) createRealTimeAlarmToModel(ctx context.Context, param *bo.CreateAlarmItemParams) (*alarmmodel.RealtimeAlarm, error) {
	strategyID := param.StrategyID
	levelID := param.LevelID
	bizQuery, err := getBizQuery(ctx, r.data)
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
	return &alarmmodel.RealtimeAlarm{
		Status:      vobj.ToAlertStatus(param.Status),
		StartsAt:    param.StartsAt,
		EndsAt:      param.EndsAt,
		Expr:        strategy.Expr,
		Fingerprint: param.Fingerprint,
		Labels:      vobj.NewLabels(param.Labels),
		Annotations: param.Annotations,
		RawInfoID:   param.RawID,
		RealtimeDetails: &alarmmodel.RealtimeDetails{
			Strategy:   strategy.String(),
			Level:      strategyLevel.String(),
			Datasource: "",
		},
	}, nil
}
