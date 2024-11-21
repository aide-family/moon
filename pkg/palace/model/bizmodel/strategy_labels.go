package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameStrategyMetricsLabelNotices = "strategy_metrics_label_notices"

// StrategyMetricsLabelNotice 策略labels表
type StrategyMetricsLabelNotice struct {
	model.AllFieldModel
	// label key
	Name string `gorm:"column:name;type:varchar(255);not null;comment:label名称;uniqueIndex:idx__level_id__name" json:"name"`
	// label value
	Value string `gorm:"column:value;type:varchar(255);not null;comment:标签值" json:"value"`
	// 策略等级ID
	LevelID uint32 `gorm:"column:level_id;type:int unsigned;not null;comment:策略等级ID;uniqueIndex:idx__level_id__name" json:"level_id"`
	// labels告警组
	AlarmGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_metrics_labels_alarm_groups;" json:"alarm_groups"`
}

// UnmarshalBinary redis存储实现
func (c *StrategyMetricsLabelNotice) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMetricsLabelNotice) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyLabelNotice 's table name
func (*StrategyMetricsLabelNotice) TableName() string {
	return tableNameStrategyMetricsLabelNotices
}
