package do

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// StrategyMetricReceiver binds receivers to a strategy (optional level_uid for StrategyMetricBindReceivers).
type StrategyMetricReceiver struct {
	BaseModel
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	NamespaceUID snowflake.ID   `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	StrategyUID  snowflake.ID   `gorm:"column:strategy_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	ReceiverUID  snowflake.ID   `gorm:"column:receiver_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	LevelUID     snowflake.ID   `gorm:"column:level_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
}

func (StrategyMetricReceiver) TableName() string {
	return "strategy_metric_receivers"
}

func (s *StrategyMetricReceiver) WithNamespace(namespace snowflake.ID) *StrategyMetricReceiver {
	s.NamespaceUID = namespace
	return s
}
