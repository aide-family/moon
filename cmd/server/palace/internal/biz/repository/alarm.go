package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// Alarm 告警相关接口定义
type Alarm interface {
	// GetRealTimeAlarm 获取实时告警
	GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*bizmodel.RealtimeAlarm, error)

	// GetRealTimeAlarms 获取实时告警列表
	GetRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*bizmodel.RealtimeAlarm, error)
}
