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
	DataType vobj.MQDataType `json:"data_type,omitempty"`
	// 条件
	Condition vobj.MQCondition `json:"condition,omitempty"`
	// 所属策略
	StrategyID uint32 `json:"strategyID,omitempty"`
	// 告警等级
	AlarmLevelID uint32   `json:"alarmLevelID,omitempty"`
	AlarmLevel   *SysDict `json:"alarmLevel,omitempty"`
	// 状态
	Status vobj.Status `json:"status,omitempty"`
	// 告警页面
	AlarmPageIds []uint32 `json:"alarm_page_ids,omitempty"`
	// 告警页面
	AlarmPage []*SysDict `json:"alarm_page"`
	// 策略告警组
	AlarmGroupIds []uint32            `json:"alarm_group_ids,omitempty"`
	AlarmGroups   []*AlarmNoticeGroup `json:"alarm_groups,omitempty"`
	// object path key
	PathKey string `json:"path_key,omitempty"`
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
