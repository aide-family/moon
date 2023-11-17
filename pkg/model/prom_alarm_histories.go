package model

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromAlarmHistory = "prom_alarm_histories"

// PromAlarmHistory mapped from table <prom_alarm_histories>
type PromAlarmHistory struct {
	query.BaseModel
	Instance   string           `gorm:"column:instance;type:varchar(64);not null;comment:instance名称" json:"instance"`                                            // node名称
	Status     int32            `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态, 报警和恢复" json:"status"`                                             // 告警消息状态, 报警和恢复
	Info       string           `gorm:"column:info;type:json;not null;comment:原始告警消息" json:"info"`                                                               // 原始告警消息
	CreatedAt  time.Time        `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                      // 创建时间
	StartAt    int64            `gorm:"column:start_at;type:bigint unsigned;not null;comment:报警开始时间" json:"start_at"`                                            // 报警开始时间
	EndAt      int64            `gorm:"column:end_at;type:bigint unsigned;not null;comment:报警恢复时间" json:"end_at"`                                                // 报警恢复时间
	Duration   int64            `gorm:"column:duration;type:bigint unsigned;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0" json:"duration"`                              // 持续时间时间戳, 没有恢复, 时间戳是0
	StrategyID uint             `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__strategy_id,priority:1;comment:规则ID, 用于查询时候" json:"strategy_id"` // 规则ID, 用于查询时候
	LevelID    uint             `gorm:"column:level_id;type:int unsigned;not null;index:idx__level_id,priority:1;comment:报警等级ID" json:"level_id"`                // 报警等级ID
	Md5        string           `gorm:"column:md5;type:char(32);not null;comment:md5" json:"md5"`                                                                // md5
	Pages      []*PromAlarmPage `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:PageID;many2many:prom_prom_alarm_page_histories" json:"pages"`
}

// TableName PromAlarmHistory's table name
func (*PromAlarmHistory) TableName() string {
	return TableNamePromAlarmHistory
}
