package microrepository

import (
	"context"

	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
)

// Strategy 微服务策略推送
type Strategy interface {
	// Push 推送策略
	Push(ctx context.Context, strategies *strategyapi.PushStrategyRequest) error
}
