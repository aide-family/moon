package do

import (
	"errors"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

type StrategyMetricLevel struct {
	BaseModel
	DeletedAt    gorm.DeletedAt         `gorm:"column:deleted_at;uniqueIndex:idx__strategy_metric_levels__namespace_uid__deleted_at__strategy_uid"`
	NamespaceUID snowflake.ID           `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_metric_levels__namespace_uid__deleted_at__strategy_uid"`
	StrategyUID  snowflake.ID           `gorm:"column:strategy_uid;uniqueIndex:idx__strategy_metric_levels__namespace_uid__deleted_at__strategy_uid"`
	LevelUID     snowflake.ID           `gorm:"column:level_uid;uniqueIndex:idx__strategy_metric_levels__namespace_uid__deleted_at__level_uid"`
	Mode         enum.SampleMode        `gorm:"column:mode;type:tinyint;default:0"`
	Condition    enum.ConditionMetric   `gorm:"column:condition;type:tinyint;default:0"`
	Values       *safety.Slice[float64] `gorm:"column:values;type:json;serializer:json"`
	DurationSec  int64                  `gorm:"column:duration_sec;default:0"`
	Status       enum.GlobalStatus      `gorm:"column:status;type:tinyint;default:0"`
}

func (StrategyMetricLevel) TableName() string {
	return "strategy_metric_levels"
}

func (s *StrategyMetricLevel) WithNamespace(namespace snowflake.ID) *StrategyMetricLevel {
	s.NamespaceUID = namespace
	return s
}

func (s *StrategyMetricLevel) BeforeCreate(tx *gorm.DB) (err error) {
	if s.NamespaceUID == 0 {
		return errors.New("namespace uid is required")
	}
	if s.StrategyUID == 0 {
		return errors.New("strategy uid is required")
	}
	if s.LevelUID == 0 {
		return errors.New("level uid is required")
	}
	return nil
}
