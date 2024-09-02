package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameStrategyLabels = "strategies_labels"

// StrategyLabels 策略labels表
type StrategyLabels struct {
	model.AllFieldModel
	// 所属策略
	Name  string `gorm:"column:name;type:varchar(255);not null;comment:label名称" json:"name"`
	Value string `gorm:"column:value;type:varchar(255);not null;comment:标签值" json:"value"`
	// labels告警组
	AlarmGroups []*AlarmGroup `gorm:"many2many:strategies_labels_alarm_groups;" json:"labelsAlarmGroups"`
}

// UnmarshalBinary redis存储实现
func (c *StrategyLabels) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyLabels) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategyLabels 's table name
func (*StrategyLabels) TableName() string {
	return tableNameStrategyLabels
}
