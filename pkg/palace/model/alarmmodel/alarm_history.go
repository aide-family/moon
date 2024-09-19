package alarmmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmHistories = "alarm_histories"

// AlarmHistory 告警历史
type AlarmHistory struct {
	model.EasyModel
	// 原始信息json
	RawInfo string `gorm:"column:raw_info;type:text;not null;comment:原始信息json"`
	// 告警消息状态
	AlertStatus vobj.AlertStatus `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态"`
	// 报警开始时间
	StartAt *types.Duration `gorm:"column:start_at;type:bigint;not null;comment:报警开始时间"`
	// 报警恢复时间
	EndAt *types.Duration `gorm:"column:end_at;type:bigint;not null;comment:报警恢复时间"`
	// 报警持续时间
	Duration *types.Duration `gorm:"column:duration;type:bigint;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0"`
	// 告警表达式
	Expr string `gorm:"column:expr;type:text;not null;comment:prom ql"`
	// 指纹
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255);not null;comment:fingerprint;uniqueIndex"`
	// 附加信息
	HistoryDetails *HistoryDetails `gorm:"foreignKey:AlarmHistoryID;comment:附加信息"`
}

// String json string
func (a *AlarmHistory) String() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (a *AlarmHistory) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

// MarshalBinary redis存储实现
func (a *AlarmHistory) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

// TableName AlarmHistory's table name
func (a *AlarmHistory) TableName() string {
	return tableNameAlarmHistories
}
