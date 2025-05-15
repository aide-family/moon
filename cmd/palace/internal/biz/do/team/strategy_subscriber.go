package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.TeamStrategySubscriber = (*StrategySubscriber)(nil)

const tableNameStrategySubscriber = "team_strategy_subscribers"

type StrategySubscriber struct {
	do.TeamModel
	StrategyID    uint32          `gorm:"column:strategy_id;type:int unsigned;not null;comment:策略id" json:"strategyID"`
	Strategy      *Strategy       `gorm:"foreignKey:StrategyID;references:ID" json:"strategy"`
	SubscribeType vobj.NoticeType `gorm:"column:subscribe_type;type:int unsigned;not null;comment:订阅类型" json:"subscribeType"`
}

// GetStrategy implements do.TeamStrategySubscriber.
func (s *StrategySubscriber) GetStrategy() do.Strategy {
	if s == nil || s.Strategy == nil {
		return nil
	}
	return s.Strategy
}

// GetStrategyID implements do.TeamStrategySubscriber.
func (s *StrategySubscriber) GetStrategyID() uint32 {
	if s == nil {
		return 0
	}
	return s.StrategyID
}

// GetSubscribeType implements do.TeamStrategySubscriber.
func (s *StrategySubscriber) GetSubscribeType() vobj.NoticeType {
	if s == nil {
		return vobj.NoticeTypeUnknown
	}
	return s.SubscribeType
}

func (s *StrategySubscriber) TableName() string {
	return tableNameStrategySubscriber
}
