package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromAlarmPage = "prom_alarm_pages"
const TableNamePromStrategyAlarmPage = "prom_strategy_alarm_pages"

// PromAlarmPage 报警页面
type PromAlarmPage struct {
	BaseModel
	Name           string          `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:报警页面名称"`
	Remark         string          `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`
	Icon           string          `gorm:"column:icon;type:varchar(1024);not null;comment:图表"`
	Color          string          `gorm:"column:color;type:varchar(64);not null;comment:tab颜色"`
	Status         vo.Status       `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态,1启用;2禁用"`
	PromStrategies []*PromStrategy `gorm:"many2many:prom_strategy_alarm_pages"`
}

// TableName PromAlarmPage's table name
func (PromAlarmPage) TableName() string {
	return TableNamePromAlarmPage
}

// PromStrategyAlarmPage 报警策略与报警页面关联表
type PromStrategyAlarmPage struct {
	PromAlarmPageID uint32 `gorm:"column:alarm_page_id;type:int unsigned;not null;index:idx__prom_alarm_page_id,priority:1;comment:报警页面ID"`
	PromStrategyID  uint32 `gorm:"column:prom_strategy_id;type:int unsigned;not null;index:idx__prom_strategy_id,priority:1;comment:报警策略ID"`
}

// TableName PromStrategyAlarmPage's table name
func (PromStrategyAlarmPage) TableName() string {
	return TableNamePromStrategyAlarmPage
}
