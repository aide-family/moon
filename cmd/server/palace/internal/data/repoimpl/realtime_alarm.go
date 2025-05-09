package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel/alarmquery"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
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

// DeleteRealTimeAlarm implements repository.Alarm.
func (r *realtimeAlarmRepositoryImpl) DeleteRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error {
	// 不继续接收也不告警
	return merr.ErrorI18nToastUnimplemented(ctx, "删除告警")
}

// MarkRealTimeAlarm implements repository.Alarm.
func (r *realtimeAlarmRepositoryImpl) MarkRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error {
	// 添加处理人， 可以被多人介入
	return merr.ErrorI18nToastUnimplemented(ctx, "告警介入")
}

// SuppressRealTimeAlarm implements repository.Alarm.
func (r *realtimeAlarmRepositoryImpl) SuppressRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error {
	// 继续接收告警， 但是不告警
	return merr.ErrorI18nToastUnimplemented(ctx, "告警抑制")
}

// UpgradeRealTimeAlarm implements repository.Alarm.
func (r *realtimeAlarmRepositoryImpl) UpgradeRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error {
	// 升级告警, 把告警发送给更高级别的人
	return merr.ErrorI18nToastUnimplemented(ctx, "告警升级")
}

func (r *realtimeAlarmRepositoryImpl) CreateRealTimeAlarm(ctx context.Context, param *bo.CreateAlarmInfoParams) error {
	alarmQuery, err := getTeamBizAlarmQuery(param.TeamID, r.data)
	if err != nil {
		return err
	}

	realtimeAlarmDoQuery := alarmQuery.RealtimeAlarm
	realtimeDetailsDoQuery := alarmQuery.RealtimeDetails
	realTimes := r.createRealTimeAlarmToModels(param)
	// 所更新的字段
	realCol := []string{
		realtimeAlarmDoQuery.Summary.ColumnName().String(),
		realtimeAlarmDoQuery.Description.ColumnName().String(),
		realtimeAlarmDoQuery.Status.ColumnName().String(),
		realtimeAlarmDoQuery.StartsAt.ColumnName().String(),
		realtimeAlarmDoQuery.EndsAt.ColumnName().String(),
		realtimeAlarmDoQuery.Expr.ColumnName().String(),
		realtimeAlarmDoQuery.Labels.ColumnName().String(),
		realtimeAlarmDoQuery.Annotations.ColumnName().String(),
	}
	detailCol := []string{
		realtimeDetailsDoQuery.Strategy.ColumnName().String(),
		realtimeDetailsDoQuery.Level.ColumnName().String(),
		realtimeDetailsDoQuery.Datasource.ColumnName().String(),
	}

	return alarmQuery.Transaction(func(tx *alarmquery.Query) error {
		for _, realTime := range realTimes {
			// 实时告警表
			if err := tx.RealtimeAlarm.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: realtimeAlarmDoQuery.Fingerprint.ColumnName().String()}},
				DoUpdates: clause.AssignmentColumns(realCol),
			}).Create(realTime); err != nil {
				log.Errorw("method", "CreateRealTimeAlarm", "error", err)
				continue
			}

			// 告警详情表
			detail := &alarmmodel.RealtimeDetails{
				Strategy:        param.Strategy.String(),
				Level:           param.Level,
				Datasource:      param.GetDatasourceMap(realTime.Labels.GetDatasourceID()),
				RealtimeAlarmID: realTime.ID,
				RealtimeAlarm:   realTime,
			}
			if err := tx.RealtimeDetails.WithContext(ctx).Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: realtimeDetailsDoQuery.RealtimeAlarmID.ColumnName().String()}},
				DoUpdates: clause.AssignmentColumns(detailCol),
			}).Create(detail); err != nil {
				log.Errorw("method", "CreateRealTimeAlarm", "error", err)
				continue
			}
		}
		return nil
	})
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
	wheres := []gen.Condition{
		alarmQuery.RealtimeAlarm.ID.Eq(params.RealtimeAlarmID),
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

	go func() {
		defer after.RecoverX()
		// 删除已经恢复的告警及关联数据
		var resolvedIds []uint32
		_ = alarmQuery.WithContext(types.CopyValueCtx(ctx)).RealtimeAlarm.Where(alarmQuery.RealtimeAlarm.Status.Eq(vobj.AlertStatusResolved.GetValue())).
			Select(alarmQuery.RealtimeAlarm.ID).Scan(&resolvedIds)
		if len(resolvedIds) > 0 {
			_, _ = alarmQuery.RealtimeAlarm.Where(alarmQuery.RealtimeAlarm.Status.Eq(vobj.AlertStatusResolved.GetValue())).Delete()
			_, _ = alarmQuery.RealtimeDetails.Where(alarmQuery.RealtimeDetails.RealtimeAlarmID.In(resolvedIds...)).Delete()
			_, _ = alarmQuery.RealtimeAlarmPage.Where(alarmQuery.RealtimeAlarmPage.RealtimeAlarmID.In(resolvedIds...)).Delete()
			_, _ = alarmQuery.RealtimeAlarmReceiver.Where(alarmQuery.RealtimeAlarmReceiver.RealtimeAlarmID.In(resolvedIds...)).Delete()
		}
	}()

	wheres := []gen.Condition{
		alarmQuery.RealtimeAlarm.Status.Eq(vobj.AlertStatusFiring.GetValue()),
	}
	if !types.TextIsNull(params.EventAtStart) && !types.TextIsNull(params.EventAtEnd) {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.StartsAt.Between(params.EventAtStart, params.EventAtEnd))
	}

	// 获取指定告警页面告警数据
	if params.AlarmPageID > 0 {
		var realtimeAlarmIDs []uint32
		if err = alarmQuery.WithContext(ctx).RealtimeAlarmPage.Where(alarmQuery.RealtimeAlarmPage.PageID.Eq(params.AlarmPageID)).
			Select(alarmQuery.RealtimeAlarmPage.RealtimeAlarmID).Group(alarmQuery.RealtimeAlarmPage.RealtimeAlarmID).
			Scan(&realtimeAlarmIDs); err != nil {
			return nil, err
		}
		wheres = append(wheres, alarmQuery.RealtimeAlarm.ID.In(realtimeAlarmIDs...))
	}
	// 获取指定人员告警数据
	if params.MyAlarm {
		// 获取我的通知告警组
		bizQuery, err := getBizQuery(ctx, r.data)
		if err != nil {
			return nil, err
		}
		var alarmNoticeGroupIDs []uint32
		if err = bizQuery.AlarmNoticeMember.WithContext(ctx).
			Where(bizQuery.AlarmNoticeMember.MemberID.Eq(middleware.GetTeamMemberID(ctx))).
			Select(bizQuery.AlarmNoticeMember.AlarmGroupID).Group(bizQuery.AlarmNoticeMember.AlarmGroupID).
			Scan(&alarmNoticeGroupIDs); err != nil {
			return nil, err
		}
		var realtimeAlarmIDs []uint32
		if err = alarmQuery.WithContext(ctx).RealtimeAlarmReceiver.
			Where(alarmQuery.RealtimeAlarmReceiver.AlarmNoticeGroupID.In(alarmNoticeGroupIDs...)).
			Select(alarmQuery.RealtimeAlarmReceiver.RealtimeAlarmID).Group(alarmQuery.RealtimeAlarmReceiver.RealtimeAlarmID).
			Scan(&realtimeAlarmIDs); err != nil {
			return nil, err
		}
		wheres = append(wheres, alarmQuery.RealtimeAlarm.ID.In(realtimeAlarmIDs...))
	}

	realtimeAlarmQuery := alarmQuery.WithContext(ctx).RealtimeAlarm.Where(wheres...)
	if realtimeAlarmQuery, err = types.WithPageQuery(realtimeAlarmQuery, params.Pagination); err != nil {
		return nil, err
	}
	return realtimeAlarmQuery.Order(alarmQuery.RealtimeAlarm.ID.Desc()).Preload(alarmQuery.RealtimeAlarm.RealtimeDetails).Find()
}

func (r *realtimeAlarmRepositoryImpl) createRealTimeAlarmToModels(param *bo.CreateAlarmInfoParams) []*alarmmodel.RealtimeAlarm {
	strategy := param.Strategy
	alarms := types.SliceTo(param.Alerts, func(alarmParam *bo.AlertItemRawParams) *alarmmodel.RealtimeAlarm {
		labels := alarmParam.Labels
		annotations := alarmParam.Annotations
		alarm := &alarmmodel.RealtimeAlarm{
			Status:          vobj.ToAlertStatus(alarmParam.Status),
			StartsAt:        alarmParam.StartsAt,
			EndsAt:          alarmParam.EndsAt,
			Summary:         annotations.GetSummary(),
			Description:     annotations.GetDescription(),
			Expr:            strategy.Expr,
			Fingerprint:     alarmParam.Fingerprint,
			Labels:          labels,
			Annotations:     annotations,
			RawInfoID:       param.GetRawInfoID(alarmParam.Fingerprint),
			RealtimeDetails: &alarmmodel.RealtimeDetails{},
			StrategyID:      strategy.GetID(),
			LevelID:         strategy.Level.GetID(),
			Receiver: types.SliceTo(param.ReceiverGroupIDs, func(id uint32) *alarmmodel.RealtimeAlarmReceiver {
				return &alarmmodel.RealtimeAlarmReceiver{AlarmNoticeGroupID: id}
			}),
			Pages: types.SliceTo(strategy.Level.GetAlarmPageList(), func(page *bizmodel.SysDict) *alarmmodel.RealtimeAlarmPage {
				return &alarmmodel.RealtimeAlarmPage{PageID: page.GetID()}
			}),
		}
		return alarm
	})
	return alarms
}
