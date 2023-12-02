package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromAlarmRealtime = "prom_alarm_realtime"

// PromAlarmRealtime 实时告警信息
type PromAlarmRealtime struct {
	query.BaseModel
	// StrategyID 发生这条告警的具体策略信息
	StrategyID uint          `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__strategy_id,priority:1;comment:策略ID"`
	Strategy   *PromStrategy `gorm:"foreignKey:StrategyID"`
	Instance   string        `gorm:"column:instance;type:varchar(64);not null;index:idx__instance,priority:1;comment:instance名称"`
	Note       string        `gorm:"column:note;type:varchar(255);not null;comment:告警内容"`
	// Status 告警状态: 1告警;2恢复
	Status int32 `gorm:"column:status;type:tinyint;not null;default:1;comment:告警状态: 1告警;2恢复"`
	// EventAt 告警时间
	EventAt int64 `gorm:"column:event_at;type:bigint;not null;comment:告警时间"`
	// 通知对象, 记录事件发生时候实际的通知人员
	BeenNotifyMembers []*PromAlarmNotifyMember `gorm:"many2many:prom_realtime_alarms_notify_members;comment:已通知成员"`
	BeenChatGroups    []*PromAlarmChatGroup    `gorm:"many2many:prom_realtime_alarms_chat_groups;comment:已通知群组"`
	NotifiedAt        int64                    `gorm:"column:notified_at;type:bigint;not null;default:0;comment:通知时间"`
	// HistoryID 对应的报警历史数据
	HistoryID uint32            `gorm:"column:history_id;type:int unsigned;not null;index:idx__history_id,priority:1;comment:历史记录ID"`
	History   *PromAlarmHistory `gorm:"foreignKey:HistoryID"`
	// Intervenes 运维介入信息
	Intervenes []*PromAlarmIntervene `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmUpgradeInfo 告警升级信息
	AlarmUpgradeInfo *PromAlarmUpgrade `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmSuppressInfo 告警抑制信息
	AlarmSuppressInfo *PromAlarmSuppress `gorm:"foreignKey:RealtimeAlarmID"`
}

// TableName 表名
func (*PromAlarmRealtime) TableName() string {
	return TableNamePromAlarmRealtime
}
