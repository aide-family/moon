package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromAlarmHistory = "prom_alarm_histories"

// PromAlarmHistory 报警历史数据
type PromAlarmHistory struct {
	query.BaseModel
	Instance   string               `gorm:"column:instance;type:varchar(64);not null;comment:instance名称;index:idx__instance"`
	Status     valueobj.AlarmStatus `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态, 报警和恢复"`
	Info       string               `gorm:"column:info;type:json;not null;comment:原始告警消息"`
	StartAt    int64                `gorm:"column:start_at;type:bigint;not null;comment:报警开始时间"`
	EndAt      int64                `gorm:"column:end_at;type:bigint;not null;comment:报警恢复时间"`
	Duration   int64                `gorm:"column:duration;type:bigint;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0"`
	StrategyID uint                 `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__strategy_id,priority:1;comment:规则ID, 用于查询时候"`
	LevelID    uint                 `gorm:"column:level_id;type:int unsigned;not null;index:idx__level_id,priority:1;comment:报警等级ID"`
	Md5        string               `gorm:"column:md5;type:char(32);not null;unique:idx__md5,priority:1;comment:md5"`

	Strategy *PromStrategy `gorm:"foreignKey:StrategyID"`
	Level    *PromDict     `gorm:"foreignKey:LevelID"`
}

// TableName PromAlarmHistory's table name
func (*PromAlarmHistory) TableName() string {
	return TableNamePromAlarmHistory
}

// GetStrategy 获取策略
func (p *PromAlarmHistory) GetStrategy() *PromStrategy {
	if p == nil {
		return nil
	}
	return p.Strategy
}

// GetLevel 获取等级
func (p *PromAlarmHistory) GetLevel() *PromDict {
	if p == nil {
		return nil
	}
	return p.Level
}
