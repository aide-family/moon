package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyPing Ping监控策略定义， 用于监控指定IP的网络延迟、丢包率等
type StrategyPing struct {
	// 告警等级ID
	LevelID uint32 `json:"level_id,omitempty"`
	// 执行频率
	Interval uint32 `json:"interval,omitempty"`
	// 策略告警组
	AlarmNoticeGroupIds []uint32 `json:"alarm_group_ids,omitempty"`
	// 策略告警组
	AlarmNoticeGroups []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
	// 总包数
	Total uint32 `json:"total,omitempty"`
	// 成功包数
	Success uint32 `json:"success,omitempty"`
	// 丢包率
	LossRate float64 `json:"loss_rate,omitempty"`
	// 平均延迟
	AvgDelay uint32 `json:"avg_delay,omitempty"`
	// 最大延迟
	MaxDelay uint32 `json:"max_delay,omitempty"`
	// 最小延迟
	MinDelay uint32 `json:"min_delay,omitempty"`
	// 标准差
	StdDev uint32 `json:"std_dev,omitempty"`
}

// String 字符串
func (s *StrategyPing) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyPing) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPing) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
