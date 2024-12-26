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

// String json string
func (c *StrategyMetricLevel) String() string {
	bs, _ := types.Marshal(c)
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
