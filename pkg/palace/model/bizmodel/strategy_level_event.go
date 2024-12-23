package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// StrategyEventLevel MQ策略等级
type StrategyEventLevel struct {
	// 值
	Value string `json:"value,omitempty"`
	// 数据类型
	DataType vobj.MQDataType `json:"dataType,omitempty"`
	// 条件
	Condition vobj.MQCondition `json:"condition,omitempty"`
	// 告警等级
	Level *SysDict `json:"level,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 告警页面
	AlarmPageList []*SysDict `json:"alarmPageList"`
	// 策略告警组
	AlarmGroupList []*AlarmNoticeGroup `json:"alarmGroupList,omitempty"`
	// object path key
	PathKey string `json:"pathKey,omitempty"`
}

// String json string
func (c *StrategyEventLevel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyEventLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyEventLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}
