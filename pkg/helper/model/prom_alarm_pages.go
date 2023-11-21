package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromAlarmPage = "prom_alarm_pages"

// PromAlarmPage mapped from table <prom_alarm_pages>
type PromAlarmPage struct {
	query.BaseModel
	Name           string              `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:报警页面名称" json:"name"` // 报警页面名称
	Remark         string              `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`                               // 描述信息
	Icon           string              `gorm:"column:icon;type:varchar(1024);not null;comment:图表" json:"icon"`                                    // 图表
	Color          string              `gorm:"column:color;type:varchar(64);not null;comment:tab颜色" json:"color"`                                 // tab颜色
	Status         int32               `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态,1启用;2禁用" json:"status"`                  // 启用状态,1启用;2禁用
	PromStrategies []*PromStrategy     `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:PromStrategyID;many2many:prom_strategy_alarm_pages" json:"prom_strategies"`
	Histories      []*PromAlarmHistory `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:HistoryID;many2many:prom_alarm_page_histories" json:"histories"`
}

// TableName PromAlarmPage's table name
func (*PromAlarmPage) TableName() string {
	return TableNamePromAlarmPage
}
