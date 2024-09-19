package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameRealtimeDetails = "realtime_details"

// RealtimeDetails mapped from table <RealtimeDetails>

type RealtimeDetails struct {
	model.EasyModel
	// 相关策略
	Strategy string `gorm:"column:strategy;type:varchar(2000);not null;comment:相关策略" json:"strategy"`
	// 策略等级
	Level string `gorm:"column:level;type:varchar(2000);not null;comment:策略等级" json:"level"`
	// 数据源
	Datasource string `gorm:"column:datasource;type:varchar(2000);not null;comment:数据源" json:"datasource"`
	// 实时告警ID
	RealtimeAlarmID uint32 `gorm:"column:realtime_alarm_id;type:int;comment:告警历史ID;uniqueIndex:idx__notice__realtime_alarm_id,priority:1" json:"realtime_alarm_id"`
	// 实时告警
	RealtimeAlarm *RealtimeAlarm `gorm:"foreignKey:RealtimeAlarmID" json:"realtime_alarm"`
}

// String json string
func (c *RealtimeDetails) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *RealtimeDetails) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *RealtimeDetails) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName RealtimeDetails's table name
func (*RealtimeDetails) TableName() string {
	return tableNameRealtimeDetails
}
