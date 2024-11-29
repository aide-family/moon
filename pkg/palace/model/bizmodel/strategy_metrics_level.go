package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategyMetricsLevel = "strategy_metrics_level"

// StrategyMetricsLevel 策略等级
type StrategyMetricsLevel struct {
	model.AllFieldModel
	// 所属策略
	StrategyID uint32    `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略ID;uniqueIndex:idx__strategy_id__level_id" json:"strategy_id"`
	Strategy   *Strategy `gorm:"foreignKey:StrategyID" json:"strategy"`
	// 持续时间
	Duration int64 `gorm:"column:duration;type:bigint(20);not null;comment:告警持续时间" json:"duration"`
	// 持续次数
	Count uint32 `gorm:"column:count;type:int unsigned;not null;comment:持续次数" json:"count"`
	// 持续事件类型
	SustainType vobj.Sustain `gorm:"column:sustain_type;type:int(11);not null;comment:持续类型" json:"sustain_type"`
	// 执行频率
	Interval int64 `gorm:"column:interval;type:bigint(20);not null;comment:执行频率" json:"interval"`
	// 条件
	Condition vobj.Condition `gorm:"column:condition;type:int;not null;comment:条件" json:"condition"`
	// 阈值
	Threshold float64 `gorm:"column:threshold;type:text;not null;comment:阈值" json:"threshold"`
	// 告警等级
	LevelID uint32   `gorm:"column:level_id;type:int unsigned;not null;comment:告警等级;uniqueIndex:idx__strategy_id__level_id" json:"level_id"`
	Level   *SysDict `gorm:"foreignKey:LevelID" json:"level"`
	// 状态
	Status vobj.Status `gorm:"column:status;type:int;not null;default:1;comment:策略状态" json:"status"`
	// 告警页面
	AlarmPage []*SysDict `gorm:"many2many:strategy_metrics_level_alarm_pages" json:"alarm_page"`
	// 策略告警组
	AlarmGroups []*AlarmNoticeGroup `gorm:"many2many:strategy_metrics_level_alarm_groups;" json:"alarm_groups"`
	// 策略labels
	LabelNotices []*StrategyMetricsLabelNotice `gorm:"foreignKey:LevelID;" json:"label_notices"`
}

// String json string
func (c *StrategyMetricsLevel) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *StrategyMetricsLevel) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyMetricsLevel) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategyMetricsLevel's table name
func (*StrategyMetricsLevel) TableName() string {
	return tableNameStrategyMetricsLevel
}
