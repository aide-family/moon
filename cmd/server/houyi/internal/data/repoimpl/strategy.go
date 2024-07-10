package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
)

func NewStrategyRepository(data *data.Data) repository.Strategy {
	return &strategyRepositoryImpl{data: data}
}

type strategyRepositoryImpl struct {
	data *data.Data
}

func (s *strategyRepositoryImpl) Save(ctx context.Context, strategy []*bo.Strategy) error {
	//TODO implement me
	panic("implement me")
}
