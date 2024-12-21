package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyMetricsLevel 策略等级
type StrategyMetricsLevel struct {
	// 持续时间
	Duration int64 `json:"duration,omitempty"`
	// 持续次数
	Count uint32 `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count,omitempty"`
	// 持续事件类型
	SustainType vobj.Sustain `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type,omitempty"`
	// 执行频率
	Interval int64 `gorm:"column:interval;type:bigint(20);not null;comment:执行频率" json:"interval,omitempty"`
	// 条件
	Condition vobj.Condition `gorm:"column:condition;type:int;not null;comment:条件" json:"condition,omitempty"`
	// 阈值
	Threshold float64 `gorm:"column:threshold;type:text;not null;comment:阈值" json:"threshold,omitempty"`
	// 告警等级
	LevelID uint32   `json:"level_id,omitempty"`
	Level   *SysDict `json:"level,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 告警页面
	AlarmPageIds []uint32   `json:"alarm_page_ids,omitempty"`
	AlarmPage    []*SysDict `json:"alarm_page,omitempty"`
	// 策略告警组
	AlarmGroupIds []uint32            `json:"alarm_group_ids"`
	AlarmGroups   []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
	// 策略labels
	LabelNoticeIds []uint32                      `json:"labelNoticeIds"`
	LabelNotices   []*StrategyMetricsLabelNotice `json:"label_notices,omitempty"`
}

// String json string
func (c *StrategyMetricsLevel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyMetricsLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMetricsLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}
