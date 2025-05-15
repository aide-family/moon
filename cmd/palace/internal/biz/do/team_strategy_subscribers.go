package do

import "github.com/aide-family/moon/cmd/palace/internal/biz/vobj"

type TeamStrategySubscriber interface {
	TeamBase
	GetStrategyID() uint32
	GetStrategy() Strategy
	GetSubscribeType() vobj.NoticeType
}
