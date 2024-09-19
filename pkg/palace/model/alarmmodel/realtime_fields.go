package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameRealtimeField = "realtime_fields"

// RealtimeFields mapped from table <RealtimeFields>

type RealtimeFields struct {
	model.EasyModel
	// 告警历史ID
	AlarmID uint32 `gorm:"column:alarm_id;type:int unsigned;not null;comment:告警历史ID"`
	// 相关策略
	Strategy string `gorm:"column:strategy;type:varchar(2000);not null;comment:相关策略"`
	// 策略等级
	Level string `gorm:"column:level;type:varchar(2000);not null;comment:策略等级"`
	// 数据源
	Datasource string `gorm:"column:datasource;type:varchar(2000);not null;comment:数据源"`
}

// String json string
func (c *RealtimeFields) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *RealtimeFields) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *RealtimeFields) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName RealtimeFields's table name
func (*RealtimeFields) TableName() string {
	return tableNameRealtimeField
}
