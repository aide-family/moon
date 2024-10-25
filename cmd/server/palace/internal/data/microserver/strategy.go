package microserver

import (
	"context"

	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/go-kratos/kratos/v2/log"
)

// NewStrategyRepository 创建策略仓库
func NewStrategyRepository(data *data.Data, houyiClient *data.HouYiConn) microrepository.Strategy {
	return &strategyRepositoryImpl{data: data, houyiClient: houyiClient}
}

type strategyRepositoryImpl struct {
	data *data.Data

	houyiClient *data.HouYiConn
}

func (s *strategyRepositoryImpl) Push(ctx context.Context, strategies *strategyapi.PushStrategyRequest) error {
	if strategies == nil {
		return nil
	}
	strategyReply, err := s.houyiClient.PushStrategy(ctx, strategies)
	if err != nil {
		return err
	}
	log.Debugw("strategyReply", strategyReply)
	return nil
}
