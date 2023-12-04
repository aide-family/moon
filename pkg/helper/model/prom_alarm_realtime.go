package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
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
	Status valueobj.AlarmStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:告警状态: 1告警;2恢复"`
	// EventAt 告警时间
	EventAt int64 `gorm:"column:event_at;type:bigint;not null;comment:告警时间"`
	// 通知对象, 记录事件发生时候实际的通知人员
	BeenNotifyMembers []*PromAlarmBeenNotifyMember    `gorm:"foreignKey:RealtimeAlarmID;comment:已通知成员"`
	BeenChatGroups    []*PromAlarmBeenNotifyChatGroup `gorm:"foreignKey:RealtimeAlarmID;comment:已通知群组"`
	NotifiedAt        int64                           `gorm:"column:notified_at;type:bigint;not null;default:0;comment:通知时间"`
	// HistoryID 对应的报警历史数据
	HistoryID uint32            `gorm:"column:history_id;type:int unsigned;not null;unique:idx__history_id,priority:1;comment:历史记录ID"`
	History   *PromAlarmHistory `gorm:"foreignKey:HistoryID"`
	// Intervenes 运维介入信息
	AlarmIntervenes []*PromAlarmIntervene `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmUpgradeInfo 告警升级信息
	AlarmUpgradeInfo *PromAlarmUpgrade `gorm:"foreignKey:RealtimeAlarmID"`
	// AlarmSuppressInfo 告警抑制信息
	AlarmSuppressInfo *PromAlarmSuppress `gorm:"foreignKey:RealtimeAlarmID"`
}

// TableName 表名
func (*PromAlarmRealtime) TableName() string {
	return TableNamePromAlarmRealtime
}

// GetStrategy 获取策略信息
func (p *PromAlarmRealtime) GetStrategy() *PromStrategy {
	if p.Strategy == nil {
		p.Strategy = &PromStrategy{}
		p.Strategy.ID = p.StrategyID
	}
	return p.Strategy
}

// GetAlarmIntervenes 获取运维介入信息
func (p *PromAlarmRealtime) GetAlarmIntervenes() []*PromAlarmIntervene {
	if p == nil {
		return nil
	}
	return p.AlarmIntervenes
}

// GetBeenNotifyMembers 获取通知对象
func (p *PromAlarmRealtime) GetBeenNotifyMembers() []*PromAlarmBeenNotifyMember {
	if p == nil {
		return nil
	}
	return p.BeenNotifyMembers
}

// GetBeenChatGroups 获取通知群组
func (p *PromAlarmRealtime) GetBeenChatGroups() []*PromAlarmBeenNotifyChatGroup {
	if p == nil {
		return nil
	}
	return p.BeenChatGroups
}

// GetHistory 获取历史记录
func (p *PromAlarmRealtime) GetHistory() *PromAlarmHistory {
	if p == nil {
		return nil
	}
	return p.History
}

// GetAlarmUpgradeInfo 获取告警升级信息
func (p *PromAlarmRealtime) GetAlarmUpgradeInfo() *PromAlarmUpgrade {
	if p == nil {
		return nil
	}
	return p.AlarmUpgradeInfo
}

// GetAlarmSuppressInfo 获取告警抑制信息
func (p *PromAlarmRealtime) GetAlarmSuppressInfo() *PromAlarmSuppress {
	if p == nil {
		return nil
	}
	return p.AlarmSuppressInfo
}
