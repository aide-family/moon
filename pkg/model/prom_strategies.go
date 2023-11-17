package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromStrategy = "prom_strategies"

// PromStrategy mapped from table <prom_strategies>
type PromStrategy struct {
	query.BaseModel
	GroupID      uint   `gorm:"column:group_id;type:int unsigned;not null;comment:所属规则组ID" json:"group_id"`                                             // 所属规则组ID
	Alert        string `gorm:"column:alert;type:varchar(64);not null;comment:规则名称" json:"alert"`                                                       // 规则名称
	Expr         string `gorm:"column:expr;type:text;not null;comment:prom ql" json:"expr"`                                                             // prom ql
	For          string `gorm:"column:for;type:varchar(64);not null;default:10s;comment:持续时间" json:"for"`                                               // 持续时间
	Labels       string `gorm:"column:labels;type:json;not null;comment:标签" json:"labels"`                                                              // 标签
	Annotations  string `gorm:"column:annotations;type:json;not null;comment:告警文案" json:"annotations"`                                                  // 告警文案
	AlertLevelID uint   `gorm:"column:alert_level_id;type:int;not null;index:idx__alart_level_id,priority:1;comment:告警等级dict ID" json:"alert_level_id"` // 告警等级dict ID
	Status       int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用" json:"status"`                                      // 启用状态: 1启用;2禁用
	Remark       string `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`

	AlarmPages []*PromAlarmPage `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:AlarmPageID;many2many:prom_strategy_alarm_pages" json:"alarm_pages"`
	Categories []*PromDict      `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:DictID;many2many:prom_strategy_categories" json:"categories"`
	AlertLevel *PromDict        `gorm:"foreignKey:AlertLevelID" json:"alert_level"`
	GroupInfo  *PromGroup       `gorm:"foreignKey:GroupID" json:"group_info"`
}

// TableName PromStrategy's table name
func (*PromStrategy) TableName() string {
	return TableNamePromStrategy
}
