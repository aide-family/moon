package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// NewAlarmBiz 创建告警相关业务逻辑
func NewAlarmBiz(alarmRepository repository.Alarm) *AlarmBiz {
	return &AlarmBiz{
		alarmRepository: alarmRepository,
	}
}

// AlarmBiz 告警相关业务逻辑
type AlarmBiz struct {
	alarmRepository repository.Alarm
}

// GetRealTimeAlarm 获取实时告警明细
func (b *AlarmBiz) GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*bizmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarm(ctx, params)
}

// ListRealTimeAlarms 获取实时告警列表
func (b *AlarmBiz) ListRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*bizmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarms(ctx, params)
}
