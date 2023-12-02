package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromGroup = "prom_strategy_groups"

// PromStrategyGroup 策略组
type PromStrategyGroup struct {
	query.BaseModel
	Name           string          `gorm:"column:name;type:varchar(64);not null;comment:规则组名称;index:idx__name,unique"`
	StrategyCount  int64           `gorm:"column:strategy_count;type:bigint;not null;comment:规则数量"`
	Status         int32           `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用"`
	Remark         string          `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`
	PromStrategies []*PromStrategy `gorm:"foreignKey:GroupID"`
	Categories     []*PromDict     `gorm:"many2many:prom_group_categories"`
}

// TableName PromStrategyGroup's table name
func (*PromStrategyGroup) TableName() string {
	return TableNamePromGroup
}
