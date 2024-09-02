package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// SubscriberStrategyParams 用户订阅策略参数
	SubscriberStrategyParams struct {
		StrategyID uint32          `json:"strategy_id"`
		NotifyType vobj.NotifyType `json:"notify_type"`
		UserID     uint32          `json:"user_id"`
	}

	// UnSubscriberStrategyParams 用户取消订阅
	UnSubscriberStrategyParams struct {
		StrategyID uint32 `json:"strategy_id"`
		UserID     uint32 `json:"user_id"`
	}

	// QueryStrategySubscriberParams 策略订阅者参数
	QueryStrategySubscriberParams struct {
		Page       types.Pagination
		StrategyID uint32          `json:"strategy_id"`
		NotifyType vobj.NotifyType `json:"notify_type"`
	}
	// QueryUserSubscriberParams 用户订阅策略列表查询参数
	QueryUserSubscriberParams struct {
		UserID     uint32          `json:"user_id"`
		NotifyType vobj.NotifyType `json:"notify_type"`
		Page       types.Pagination
	}
)
