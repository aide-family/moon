package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyMetricsLabelNotice 策略labels表
type StrategyMetricsLabelNotice struct {
	// label key
	Name string `json:"name"`
	// label value
	Value string `json:"value"`
	// labels告警组
	AlarmGroups []*AlarmNoticeGroup `json:"alarm_groups"`
}

// String json 序列化实现
func (c *StrategyMetricsLabelNotice) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyMetricsLabelNotice) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMetricsLabelNotice) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}
