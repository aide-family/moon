package model

import (
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameStrategyTemplateCategories = "strategy_template_categories"

// StrategyTemplateCategories 策略模板类型
type StrategyTemplateCategories struct {
	BaseModel
	StrategyTemplateID uint32 `gorm:"primaryKey"`
	SysDictID          uint32 `gorm:"primaryKey"`
}

// String json string
func (c *StrategyTemplateCategories) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary 实现redis数据转换
func (c *StrategyTemplateCategories) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary 实现redis数据转换
func (c *StrategyTemplateCategories) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyTemplateCategories's table name
func (*StrategyTemplateCategories) TableName() string {
	return tableNameStrategyTemplateCategories
}
