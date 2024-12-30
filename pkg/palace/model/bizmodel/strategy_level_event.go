package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyEventLevel 事件策略等级
type StrategyEventLevel struct {
	// 值
	Value string `json:"value,omitempty"`
	// 数据类型
	DataType vobj.EventDataType `json:"dataType,omitempty"`
	// 条件
	Condition vobj.EventCondition `json:"condition,omitempty"`
	// object path key
	PathKey string `json:"pathKey,omitempty"`

	// 告警等级ID
	Level *SysDict `json:"level,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList,omitempty"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
}

// GetLevel 获取告警等级
func (s *StrategyEventLevel) GetLevel() *SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.Level
}

// GetAlarmPageList 获取告警页面
func (s *StrategyEventLevel) GetAlarmPageList() []*SysDict {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmPageList
}

// GetAlarmGroupList 获取告警组
func (s *StrategyEventLevel) GetAlarmGroupList() []*AlarmNoticeGroup {
	if types.IsNil(s) {
		return nil
	}
	return s.AlarmGroupList
}

// String json string
func (s *StrategyEventLevel) String() string {
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
func (s *StrategyEventLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, s)
}

// MarshalBinary redis存储实现
func (s *StrategyEventLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(s)
}
