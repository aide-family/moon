package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategySubscribers = "strategy_subscribers"

// StrategySubscriber 策略订阅者信息
type StrategySubscriber struct {
	AllFieldModel
	Strategy        *Strategy       `gorm:"foreignKey:StrategyID" json:"strategy"`
	AlarmNoticeType vobj.NotifyType `gorm:"column:notice_type;type:int;not null;comment:通知类型;" json:"alarm_notice_type"`
	UserID          uint32          `gorm:"column:user_id;type:int;not null;comment:订阅人id;uniqueIndex:idx__strategy_subscriber_user_id,priority:1" json:"user_id"`
	StrategyID      uint32          `gorm:"column:strategy_id;type:int;comment:告警分组id;uniqueIndex:idx__strategy_subscriber_user_id,priority:2" json:"strategy_id"`
}

// UnmarshalBinary redis存储实现
func (c *StrategySubscriber) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategySubscriber) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName StrategySubscriber  table name
func (*StrategySubscriber) TableName() string {
	return tableNameStrategySubscribers
}

// GetStrategy get strategy
func (c *StrategySubscriber) GetStrategy() *Strategy {
	if types.IsNil(c) {
		return nil
	}
	return c.Strategy
}

// GetAlarmNoticeType get alarm notice type
func (c *StrategySubscriber) GetAlarmNoticeType() vobj.NotifyType {
	if types.IsNil(c) {
		return 0
	}
	return c.AlarmNoticeType
}

// GetUserID get user id
func (c *StrategySubscriber) GetUserID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.UserID
}

// GetStrategyID get strategy id
func (c *StrategySubscriber) GetStrategyID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.StrategyID
}
