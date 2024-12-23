package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

// StrategyPortLevel 策略端口
type StrategyPortLevel struct {
	// 告警等级I
	Level *SysDict `json:"level,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarmGroupList,omitempty"`
	// 阈值
	Threshold int64 `json:"threshold,omitempty"`
	// 端口
	Port uint32 `json:"port,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
}

// String 字符串
func (s *StrategyPortLevel) String() string {
	bs, _ := types.Marshal(s)
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
