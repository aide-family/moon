package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyPortLevel 策略端口
type StrategyPortLevel struct {
	// 阈值
	Threshold int64 `json:"threshold"`
	// 端口
	Port uint32 `json:"port"`

	// 告警等级I
	Level *SysDict `json:"level,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarmGroupList,omitempty"`
}

// GetLevel 获取告警等级
func (s *StrategyPortLevel) GetLevel() *SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.Level
}

// GetAlarmPageList 获取告警页面
func (s *StrategyPortLevel) GetAlarmPageList() []*SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmPageList
}

// GetAlarmGroupList 获取告警组
func (s *StrategyPortLevel) GetAlarmGroupList() []*AlarmNoticeGroup {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmGroupList
}

// String 字符串
func (s *StrategyPortLevel) String() string {
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
func (s *StrategyPortLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyPortLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
