package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategyMqLevel = "strategy_mq_level"

// StrategyMQLevel MQ策略等级
type StrategyMQLevel struct {
	model.AllFieldModel
	// 值
	Value string `gorm:"column:value;type:varchar(255);not null;comment:MQ值" json:"value"`
	// 数据类型
	DataType vobj.MQDataType `gorm:"column:data_type;type:int;not null;comment:数据类型" json:"data_type"`
	// 条件
	Condition vobj.MQCondition `gorm:"column:condition;type:int;not null;comment:条件" json:"condition"`
	// 所属策略
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__strategy_id__mq_level_id" json:"strategyID"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID" json:"strategy"`
	// 告警等级
	AlarmLevelID uint32   `gorm:"column:alarm_level_id;type:int unsigned;not null;comment:告警等级;uniqueIndex:idx__strategy_id__alarm_level_id" json:"alarmLevelID"`
	AlarmLevel   *SysDict `gorm:"foreignKey:AlarmLevelID" json:"alarmLevel"`
	// 状态
	Status vobj.Status `gorm:"column:status;type:int;not null;default:1;comment:策略状态" json:"status"`
	// 告警页面
	AlarmPage []*SysDict `gorm:"many2many:strategy_mq_level_alarm_pages" json:"alarm_page"`
	// 策略告警组
	AlarmGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_mq_level_alarm_groups;" json:"alarm_groups"`
	// object path key
	PathKey string `gorm:"column:path_key;path_key:varchar(255);not null;comment:objectPathKey" json:"path_key"`
}

// String json string
func (c *StrategyMQLevel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyMQLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMQLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyMQLevel's table name
func (*StrategyMQLevel) TableName() string {
	return tableNameStrategyMqLevel
}
