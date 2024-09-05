package microrepository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
)

// Strategy 微服务策略推送
type Strategy interface {
	// Push 推送策略
	Push(ctx context.Context, strategies []*bo.Strategy) error
}
