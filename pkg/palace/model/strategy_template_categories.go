package model

import "encoding/json"

const TableNameStrategyTemplateCategories = "strategy_template_categories"

type StrategyTemplateCategories struct {
	BaseModel
	StrategyTemplateID uint32 `gorm:"primaryKey"`
	SysDictID          uint32 `gorm:"primaryKey"`
}

// String json string
func (c *StrategyTemplateCategories) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyTemplateCategories) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyTemplateCategories) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyTemplateCategories's table name
func (*StrategyTemplateCategories) TableName() string {
	return TableNameStrategyTemplateCategories
}
