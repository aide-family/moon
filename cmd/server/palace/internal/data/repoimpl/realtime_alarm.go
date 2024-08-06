package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen"
	"gorm.io/gorm"
)

// NewRealtimeAlarmRepository 实例化告警业务数据库
func NewRealtimeAlarmRepository(data *data.Data) repository.Alarm {
	return &realtimeAlarmRepositoryImpl{data: data}
}

type realtimeAlarmRepositoryImpl struct {
	data *data.Data
}

// getBizQuery 获取告警业务数据库
func getBizAlarmQuery(ctx context.Context, data *data.Data) (*bizquery.Query, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	bizDB, err := data.GetAlarmGormDB(claims.GetTeam())
	if !types.IsNil(err) {
		return nil, err
	}
	return bizquery.Use(bizDB), nil
}

func (r *realtimeAlarmRepositoryImpl) GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*bizmodel.RealtimeAlarm, error) {
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
	detail, err := alarmQuery.WithContext(ctx).RealtimeAlarm.Where(wheres...).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nAlarmDataNotFoundErr(ctx).WithCause(err)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return detail, nil
}

func (r *realtimeAlarmRepositoryImpl) GetRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*bizmodel.RealtimeAlarm, error) {
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
	if len(params.AlarmLevels) > 0 {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.LevelID.In(params.AlarmLevels...))
	}
	if len(params.AlarmStatuses) > 0 {
		statuses := types.SliceTo(params.AlarmStatuses, func(status vobj.AlertStatus) int {
			return status.GetValue()
		})
		wheres = append(wheres, alarmQuery.RealtimeAlarm.Status.In(statuses...))
	}
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, alarmQuery.RealtimeAlarm.RawInfo.Like(params.Keyword))
	}
	// TODO 获取指定告警页面告警数据
	// TODO 获取指定人员告警数据
	realtimeAlarmQuery := alarmQuery.WithContext(ctx).RealtimeAlarm.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IRealtimeAlarmDo](realtimeAlarmQuery, params.Pagination); err != nil {
		return nil, err
	}
	return realtimeAlarmQuery.Find()
}
