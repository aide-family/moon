package bizmodel

import "github.com/aide-family/moon/pkg/util/types"

const tableNameStrategyCategories = "strategy_categories"

// StrategyCategories 策略类型中间表
type StrategyCategories struct {
	StrategyID uint32 `gorm:"primaryKey;column:strategy_id;type:int unsigned;primaryKey" json:"strategy_id"`
	SysDictID  uint32 `gorm:"primaryKey;column:sys_dict_id;type:int unsigned;primaryKey" json:"sys_dict_id"`
}

// UnmarshalBinary redis存储实现
func (c *StrategyCategories) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyCategories) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyCategories 's table name
func (*StrategyCategories) TableName() string {
	return tableNameStrategyCategories
}
