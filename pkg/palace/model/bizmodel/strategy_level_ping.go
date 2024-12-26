package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyPingLevel Ping监控策略定义， 用于监控指定IP的网络延迟、丢包率等
type StrategyPingLevel struct {
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

	// 告警等级ID
	Level *SysDict `json:"level,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
}

// GetLevel 获取告警等级
func (s *StrategyPingLevel) GetLevel() *SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.Level
}

// GetAlarmPageList 获取告警页面
func (s *StrategyPingLevel) GetAlarmPageList() []*SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmPageList
}

// GetAlarmGroupList 获取告警组
func (s *StrategyPingLevel) GetAlarmGroupList() []*AlarmNoticeGroup {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmGroupList
}

// String 字符串
func (s *StrategyPingLevel) String() string {
	if s == nil {
		return "{}"
	}
	bs, err := types.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (s *StrategyPingLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPingLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
