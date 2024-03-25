package do

import (
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNamePromStrategyNotifyTemplate = "prom_strategy_notify_templates"

const (
	PromStrategyNotifyTemplateFieldStrategyID = "strategy_id"
)

func PromStrategyNotifyTemplateWhereStrategyID(strategyID uint32) basescopes.ScopeMethod {
	return basescopes.WhereInColumn(PromStrategyNotifyTemplateFieldStrategyID, strategyID)
}

type PromStrategyNotifyTemplate struct {
	BaseModel
	StrategyID uint32                  `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID"`
	NotifyType vobj.NotifyTemplateType `gorm:"column:notify_type;type:tinyint;not null;comment:通知类型"`
	Content    string                  `gorm:"column:content;type:text;not null;comment:通知内容模板"`
}

// TableName 表名
func (*PromStrategyNotifyTemplate) TableName() string {
	return TableNamePromStrategyNotifyTemplate
}
