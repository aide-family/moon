package do

import (
	"errors"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// StrategyMetric one row per strategy (SaveStrategyMetric upserts by strategy_uid).
type StrategyMetric struct {
	BaseModel
	DeletedAt      gorm.DeletedAt              `gorm:"column:deleted_at;uniqueIndex:idx__strategy_metrics__namespace_uid__deleted_at__strategy_uid"`
	NamespaceUID   snowflake.ID                `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_metrics__namespace_uid__deleted_at__strategy_uid"`
	StrategyUID    snowflake.ID                `gorm:"column:strategy_uid;uniqueIndex:idx__strategy_metrics__namespace_uid__deleted_at__strategy_uid"`
	Expr           string                      `gorm:"column:expr;type:text;default:''"`
	Labels         *safety.Map[string, string] `gorm:"column:labels;type:json;"`
	Summary        string                      `gorm:"column:summary;type:varchar(500);default:''"`
	Description    string                      `gorm:"column:description;type:text;default:''"`
	Status         enum.GlobalStatus           `gorm:"column:status;type:tinyint;default:0"`
	DatasourceUIDs *safety.Slice[int64]        `gorm:"column:datasource_uids;type:json;serializer:json"`
}

func (StrategyMetric) TableName() string {
	return "strategy_metrics"
}

func (s *StrategyMetric) WithNamespace(namespace snowflake.ID) *StrategyMetric {
	s.NamespaceUID = namespace
	return s
}

func (s *StrategyMetric) BeforeCreate(tx *gorm.DB) (err error) {
	if s.NamespaceUID == 0 {
		return errors.New("namespace uid is required")
	}
	if s.StrategyUID == 0 {
		return errors.New("strategy uid is required")
	}
	if s.Expr == "" {
		return errors.New("expr is required")
	}
	return nil
}
