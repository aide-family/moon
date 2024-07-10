package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
)

func NewStrategyBiz(strategyRepository repository.Strategy) *StrategyBiz {
	return &StrategyBiz{
		strategyRepository: strategyRepository,
	}
}

// StrategyBiz .
type StrategyBiz struct {
	strategyRepository repository.Strategy
}

// SaveStrategy 保存策略信息
func (s *StrategyBiz) SaveStrategy(ctx context.Context, strategy []*bo.Strategy) error {
	return s.strategyRepository.Save(ctx, strategy)
}
