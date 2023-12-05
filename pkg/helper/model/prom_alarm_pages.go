package model

import (
	query "github.com/aide-cloud/gorm-normalize"

	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromAlarmPage = "prom_alarm_pages"

// PromAlarmPage 报警页面
type PromAlarmPage struct {
	query.BaseModel
	Name               string               `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:报警页面名称"`
	Remark             string               `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`
	Icon               string               `gorm:"column:icon;type:varchar(1024);not null;comment:图表"`
	Color              string               `gorm:"column:color;type:varchar(64);not null;comment:tab颜色"`
	Status             valueobj.Status      `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态,1启用;2禁用"`
	PromStrategies     []*PromStrategy      `gorm:"many2many:prom_strategy_alarm_pages"`
	Histories          []*PromAlarmHistory  `gorm:"many2many:prom_alarm_page_histories"`
	PromRealtimeAlarms []*PromAlarmRealtime `gorm:"many2many:prom_alarm_page_realtime_alarms"`
}

// TableName PromAlarmPage's table name
func (*PromAlarmPage) TableName() string {
	return TableNamePromAlarmPage
}
