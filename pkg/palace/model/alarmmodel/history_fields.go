package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameHistoryField = "history_fields"

// HistoryFields 告警历史字段表
type HistoryFields struct {
	model.EasyModel
	// 告警历史ID
	AlarmID uint32 `gorm:"column:alarm_id;type:int;comment:告警历史ID;uniqueIndex:idx__notice__alarm_history_id,priority:1" json:"alarm_id"`
	// 相关策略
	Strategy string `gorm:"column:strategy;type:varchar(2000);not null;comment:相关策略"`
	// 策略等级
	Level string `gorm:"column:level;type:varchar(2000);not null;comment:策略等级"`
	// 数据源
	Datasource string `gorm:"column:datasource;type:varchar(2000);not null;comment:数据源"`
}

// UnmarshalBinary redis存储实现
func (c *HistoryFields) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *HistoryFields) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName HistoryFields's table name
func (*HistoryFields) TableName() string {
	return tableNameHistoryField
}
