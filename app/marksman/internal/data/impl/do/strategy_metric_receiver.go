package do

import (
	"errors"
	"time"

	"github.com/aide-family/magicbox/contextx"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// StrategyMetricReceiver binds receivers to a strategy (optional level_uid for StrategyMetricBindReceivers).
type StrategyMetricReceiver struct {
	ID           uint32         `gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt    time.Time      `gorm:"column:created_at;"`
	Creator      snowflake.ID   `gorm:"column:creator;index"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	NamespaceUID snowflake.ID   `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	StrategyUID  snowflake.ID   `gorm:"column:strategy_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	ReceiverUID  snowflake.ID   `gorm:"column:receiver_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
	LevelUID     snowflake.ID   `gorm:"column:level_uid;uniqueIndex:idx__strategy_metric_receivers__namespace_uid__deleted_at__strategy_uid__level_uid__receiver_uid"`
}

func (StrategyMetricReceiver) TableName() string {
	return "strategy_metric_receivers"
}

func (s *StrategyMetricReceiver) BeforeCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	if s.Creator == 0 {
		s.Creator = contextx.GetUserUID(ctx)
	}
	if s.NamespaceUID == 0 {
		s.NamespaceUID = contextx.GetNamespace(ctx)
	}
	if s.StrategyUID == 0 {
		return errors.New("strategy uid is required")
	}
	if s.ReceiverUID == 0 {
		return errors.New("receiver uid is required")
	}
	return nil
}
