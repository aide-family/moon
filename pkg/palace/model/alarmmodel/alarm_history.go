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
	// 实例名称
	InstanceName string `gorm:"column:alert;type:varchar(64);not null;comment:实例名称" json:"instanceName"`
	// 告警消息状态
	AlertStatus vobj.AlertStatus `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态"`
	// 原始告警消息
	Info string `gorm:"column:info;type:json;not null;comment:原始告警消息"`
	// 报警开始时间
	StartAt *types.Duration `gorm:"column:start_at;type:bigint;not null;comment:报警开始时间"`
	// 报警恢复时间
	EndAt *types.Duration `gorm:"column:end_at;type:bigint;not null;comment:报警恢复时间"`
	// 报警持续时间
	Duration *types.Duration `gorm:"column:duration;type:bigint;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0"`
	// 策略ID
	StrategyID uint32 `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID" json:"strategy_id"`
	// 告警等级id
	LevelID uint32 `gorm:"column:level_id;type:int unsigned;not null;index:idx__h__level_id,priority:1;comment:报警等级ID"`
	// 告警表达式
	Expr string `gorm:"column:expr;type:text;not null;comment:prom ql"`
	// 数据源
	DatasourceID uint32 `gorm:"column:datasource_id;type:int unsigned;not null;comment:数据源ID"`
	// 指纹
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255);not null;comment:fingerprint;uniqueIndex"`
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
