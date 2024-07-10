package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
)

// Strategy .
type Strategy interface {
	// Save 保存策略
	Save(ctx context.Context, strategy []*bo.Strategy) error
}
