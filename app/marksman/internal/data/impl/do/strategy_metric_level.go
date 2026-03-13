package do

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type StrategyMetricLevel struct {
	BaseModel
	DeletedAt    gorm.DeletedAt         `gorm:"column:deleted_at;uniqueIndex:idx__strategy_metric_levels__ns__deleted__strategy__level"`
	NamespaceUID snowflake.ID           `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_metric_levels__ns__deleted__strategy__level"`
	StrategyUID  snowflake.ID           `gorm:"column:strategy_uid;uniqueIndex:idx__strategy_metric_levels__ns__deleted__strategy__level"`
	LevelUID     snowflake.ID           `gorm:"column:level_uid;uniqueIndex:idx__strategy_metric_levels__ns__deleted__strategy__level"`
	Mode         enum.SampleMode        `gorm:"column:mode;type:tinyint;default:0"`
	Condition    enum.ConditionMetric   `gorm:"column:condition;type:tinyint;default:0"`
	Values       *safety.Slice[float64] `gorm:"column:values;type:json;"`
	DurationSec  int64                  `gorm:"column:duration_sec;default:0"`
	Status       enum.GlobalStatus      `gorm:"column:status;type:tinyint;default:0"`
	Level        *Level                 `gorm:"foreignKey:LevelUID;references:ID"`
}

func (StrategyMetricLevel) TableName() string {
	return "strategy_metric_levels"
}

func (s *StrategyMetricLevel) WithNamespace(namespace snowflake.ID) *StrategyMetricLevel {
	s.NamespaceUID = namespace
	return s
}
