package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const TableNameStrategyAlarmLevel = "strategy_alarm_levels"

// StrategyAlarmLevel mapped from table <strategy_levels>
type StrategyAlarmLevel struct {
	AllFieldModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__level__name,priority:1;comment:api名称" json:"name"` // api名称
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                            // 状态
	Level  int         `gorm:"column:level;type:int;not null;comment:告警等级" json:"level"`
	Color  string      `gorm:"column:color;type:varchar(64);not null;comment:颜色" json:"color"`
}

// String json string
func (c *StrategyAlarmLevel) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *StrategyAlarmLevel) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *StrategyAlarmLevel) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyAlarmLevel's table name
func (*StrategyAlarmLevel) TableName() string {
	return TableNameStrategyAlarmLevel
}
