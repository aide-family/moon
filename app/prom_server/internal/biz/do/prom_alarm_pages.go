package do

import (
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
)

const TableNamePromStrategyAlarmPage = "prom_strategy_alarm_pages"

const (
	PromStrategyAlarmPageFieldPromStrategyID = "prom_strategy_id"
	PromStrategyAlarmPageFieldAlarmPageID    = "sys_dict_id"
)

// StrategyInAlarmPageIds in alarm_page_ids
func StrategyInAlarmPageIds(ids ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromStrategyAlarmPageFieldAlarmPageID, ids...)
}

// StrategyInStrategyIds in prom_strategy_id
func StrategyInStrategyIds(ids ...uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromStrategyAlarmPageFieldPromStrategyID, ids...)
}

// PromStrategyAlarmPage 报警策略与报警页面关联表
type PromStrategyAlarmPage struct {
	PromAlarmPageID uint32 `gorm:"column:sys_dict_id;type:int unsigned;not null;index:idx__prom_alarm_page_id,priority:1;comment:报警页面ID"`
	PromStrategyID  uint32 `gorm:"column:prom_strategy_id;type:int unsigned;not null;index:idx__prom_strategy_id,priority:1;comment:报警策略ID"`
}

// TableName PromStrategyAlarmPage's table name
func (PromStrategyAlarmPage) TableName() string {
	return TableNamePromStrategyAlarmPage
}
