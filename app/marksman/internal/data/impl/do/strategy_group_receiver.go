package do

import (
	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

// StrategyGroupReceiver binds receivers to a strategy group (replace semantics in BindReceivers).
type StrategyGroupReceiver struct {
	BaseModel
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;uniqueIndex:idx__strategy_group_receivers__namespace_uid__deleted_at__receiver_uid"`
	NamespaceUID     snowflake.ID   `gorm:"column:namespace_uid;uniqueIndex:idx__strategy_group_receivers__namespace_uid__deleted_at__receiver_uid"`
	StrategyGroupUID snowflake.ID   `gorm:"column:strategy_group_uid;uniqueIndex:idx__strategy_group_receivers__namespace_uid__deleted_at__receiver_uid"`
	ReceiverUID      snowflake.ID   `gorm:"column:receiver_uid;uniqueIndex:idx__strategy_group_receivers__namespace_uid__deleted_at__receiver_uid"`
}

func (StrategyGroupReceiver) TableName() string {
	return "strategy_group_receivers"
}

func (s *StrategyGroupReceiver) WithNamespace(namespace snowflake.ID) *StrategyGroupReceiver {
	s.NamespaceUID = namespace
	return s
}
