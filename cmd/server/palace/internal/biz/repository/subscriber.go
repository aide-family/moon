package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// SubscriberStrategy 订阅仓库接口
	SubscriberStrategy interface {
		// UserSubscriberStrategy 用户订阅策略
		UserSubscriberStrategy(ctx context.Context, params *bo.SubscriberStrategyParams) error
		// UserUnSubscriberStrategy 用户取消订阅策略
		UserUnSubscriberStrategy(ctx context.Context, params *bo.UnSubscriberStrategyParams) error
		// UserSubscriberStrategyList 用户订阅策略列表
		UserSubscriberStrategyList(ctx context.Context, params *bo.QueryUserSubscriberParams) ([]*bizmodel.StrategySubscribers, error)
		// StrategySubscriberList 策略订阅用户列表
		StrategySubscriberList(ctx context.Context, params *bo.QueryStrategySubscriberParams) ([]*bizmodel.StrategySubscribers, error)
	}
)
