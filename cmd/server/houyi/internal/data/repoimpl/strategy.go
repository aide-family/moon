package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/go-kratos/kratos/v2/log"
)

func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{data: data}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

func (s *strategyRepositoryImpl) Save(_ context.Context, strategies []*bo.Strategy) error {
	queue := s.data.GetStrategyQueue()
	go func() {
		defer after.RecoverX()
		for _, strategyItem := range strategies {
			if err := queue.Push(strategyItem.Message()); err != nil {
				log.Errorw("method", "queue.push", "error", err)
			}
		}
	}()

	return nil
}
