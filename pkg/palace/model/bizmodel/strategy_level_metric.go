package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyMetricLevel 策略等级
type StrategyMetricLevel struct {
	// 持续时间
	Duration int64 `json:"duration,omitempty"`
	// 持续次数
	Count uint32 `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count,omitempty"`
	// 持续事件类型
	SustainType vobj.Sustain `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type,omitempty"`
	// 条件
	Condition vobj.Condition `gorm:"column:condition;type:int;not null;comment:条件" json:"condition,omitempty"`
	// 阈值
	Threshold float64 `gorm:"column:threshold;type:text;not null;comment:阈值" json:"threshold"`

	Level           *SysDict                      `json:"level,omitempty"`
	AlarmPageList   []*SysDict                    `json:"alarm_page,omitempty"`
	AlarmGroupList  []*AlarmNoticeGroup           `json:"alarm_groups,omitempty"`
	LabelNoticeList []*StrategyMetricsLabelNotice `json:"label_notices,omitempty"`
}

// GetLevel 获取告警等级
func (s *StrategyMetricLevel) GetLevel() *SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.Level
}

// GetAlarmPageList 获取告警页面
func (s *StrategyMetricLevel) GetAlarmPageList() []*SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmPageList
}

// GetAlarmGroupList 获取告警组
func (s *StrategyMetricLevel) GetAlarmGroupList() []*AlarmNoticeGroup {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmGroupList
}

// GetLabelNoticeList 获取标签告警
func (s *StrategyMetricLevel) GetLabelNoticeList() []*StrategyMetricsLabelNotice {
	if types.IsNil(s) {
		return nil
	}
	return s.LabelNoticeList
}

// String json string
func (c *StrategyMetricLevel) String() string {
	if c == nil {
		return "{}"
	}
	bs, err := types.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyMetricLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMetricLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}
