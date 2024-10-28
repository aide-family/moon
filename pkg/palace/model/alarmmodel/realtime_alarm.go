package alarmmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameRealtimeAlarm = "realtime_alarm"

// RealtimeAlarm mapped from table <RealtimeAlarm>
type RealtimeAlarm struct {
	model.EasyModel
	// 告警状态: 1告警;2恢复
	Status vobj.AlertStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:告警状态: 1告警;2恢复"`
	// 告警时间
	StartsAt string `gorm:"column:starts_at;type:varchar(100);not null;comment:告警时间"`
	// 恢复时间
	EndsAt string `gorm:"column:ends_at;type:varchar(100);not null;comment:恢复时间"`
	// 告警摘要
	Summary string `gorm:"column:summary;type:text;not null;comment:告警摘要"`
	// 告警明细
	Description string `gorm:"column:description;type:text;not null;comment:告警明细"`
	// 触发告警表达式
	Expr string `gorm:"column:expr;type:text;not null;comment:告警表达式"`
	// 指纹
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255);not null;comment:fingerprint;uniqueIndex"`
	// 标签
	Labels *vobj.Labels `gorm:"column:labels;type:JSON;not null;comment:标签" json:"labels"`
	// 注解
	Annotations vobj.Annotations `gorm:"column:annotations;type:JSON;not null;comment:注解" json:"annotations"`
	// 告警原始数据ID
	RawInfoID uint32 `gorm:"column:raw_info_id;type:int;comment:告警原始数据id;uniqueIndex:idx__realtime__notice__raw_info_id,priority:1" json:"rawInfoId"`
	// 实时告警详情
	RealtimeDetails *RealtimeDetails `gorm:"foreignKey:RealtimeAlarmID"`
	// 策略ID
	StrategyID uint32 `gorm:"column:strategy_id;type:int;not null;comment:策略id"`
	// 告警等级ID
	LevelID uint32 `gorm:"column:level_id;type:int;not null;comment:告警等级id"`
	// 告警原始信息
	RawInfo *AlarmRaw `gorm:"foreignKey:RawInfoID"`
}

// GetRawInfo 获取告警原始信息
func (c *RealtimeAlarm) GetRawInfo() *AlarmRaw {
	if types.IsNil(c) || types.IsNil(c.RawInfo) {
		return &AlarmRaw{}
	}
	return c.RawInfo
}

// String json string
func (c *RealtimeAlarm) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *RealtimeAlarm) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *RealtimeAlarm) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName RealtimeAlarm's table name
func (*RealtimeAlarm) TableName() string {
	return tableNameRealtimeAlarm
}

// GetRealtimeDetails 获取实时告警详情
func (c *RealtimeAlarm) GetRealtimeDetails() *RealtimeDetails {
	if types.IsNil(c) || types.IsNil(c.RealtimeDetails) {
		return &RealtimeDetails{}
	}
	return c.RealtimeDetails
}
