package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

// Alarm 告警相关接口定义
type Alarm interface {
	// GetRealTimeAlarm 获取实时告警
	GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*alarmmodel.RealtimeAlarm, error)

	// GetRealTimeAlarms 获取实时告警列表
	GetRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*alarmmodel.RealtimeAlarm, error)

	// SaveAlertQueue 保存告警队列
	SaveAlertQueue(param *bo.CreateAlarmHookRawParams) error

	// CreateRealTimeAlarm 创建实时告警
	CreateRealTimeAlarm(ctx context.Context, param *bo.CreateAlarmInfoParams) error

	// MarkRealTimeAlarm 告警标记
	MarkRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error

	// DeleteRealTimeAlarm 删除实时告警
	DeleteRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error

	// SuppressRealTimeAlarm 抑制实时告警
	SuppressRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error

	// UpgradeRealTimeAlarm 升级实时告警
	UpgradeRealTimeAlarm(ctx context.Context, params *bo.MarkRealTimeAlarmParams) error
}
