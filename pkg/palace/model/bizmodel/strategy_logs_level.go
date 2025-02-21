package bizmodel

import "github.com/aide-family/moon/pkg/util/types"

// StrategyLogsLevel 策略告警日志等级
type StrategyLogsLevel struct {
	// 持续时间
	Duration int64 `json:"duration,omitempty"`
	// 持续次数
	Count uint32 `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count,omitempty"`
	// 告警等级ID
	Level *SysDict `json:"level,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
}

// GetDuration 获取持续时间
func (s *StrategyLogsLevel) GetDuration() int64 {
	if types.IsNil(s) {
		return 0
	}
	return s.Duration
}

// GetCount 获取持续次数
func (s *StrategyLogsLevel) GetCount() uint32 {
	if types.IsNil(s) {
		return 0
	}
	return s.Count
}

// GetLevel 获取告警等级
func (s *StrategyLogsLevel) GetLevel() *SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.Level
}

// GetAlarmPageList 获取告警页面
func (s *StrategyLogsLevel) GetAlarmPageList() []*SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmPageList
}

// GetAlarmGroupList 获取告警组
func (s *StrategyLogsLevel) GetAlarmGroupList() []*AlarmNoticeGroup {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmGroupList
}

// UnmarshalBinary redis存储实现
func (s *StrategyLogsLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyLogsLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}

// String 返回字符串
func (s *StrategyLogsLevel) String() string {
	if s == nil {
		return "{}"
	}
	bs, err := types.Marshal(s)
	if err != nil {
		return "{}"
	}
	return string(bs)
}
