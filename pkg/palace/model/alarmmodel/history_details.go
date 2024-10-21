package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameHistoryDetails = "history_details"

// HistoryDetails 告警历史字段表
type HistoryDetails struct {
	model.EasyModel
	// 相关策略
	Strategy string `gorm:"column:strategy;type:text;not null;comment:相关策略" json:"strategy"`
	// 策略等级
	Level string `gorm:"column:level;type:text;not null;comment:策略等级" json:"level"`
	// 数据源
	Datasource string `gorm:"column:datasource;type:text;not null;comment:数据源" json:"datasource"`
	// 告警历史ID
	AlarmHistoryID uint32 `gorm:"column:alarm_history_id;type:int;comment:告警历史ID;uniqueIndex:idx__notice__alarm_history_id,priority:1" json:"alarmHistoryId"`

	AlarmHistory *AlarmHistory `gorm:"foreignKey:AlarmHistoryID" json:"alarm_history"`
}

// UnmarshalBinary redis存储实现
func (c *HistoryDetails) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *HistoryDetails) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName HistoryDetails's table name
func (*HistoryDetails) TableName() string {
	return tableNameHistoryDetails
}
