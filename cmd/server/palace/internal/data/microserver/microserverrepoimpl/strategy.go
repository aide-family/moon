package microserverrepoimpl

import (
	"context"

	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/go-kratos/kratos/v2/log"
)

func NewStrategyRepository(data *data.Data, houyiClient *microserver.HouYiConn) microrepository.Strategy {
	return &strategyRepositoryImpl{data: data, houyiClient: houyiClient}
}

type strategyRepositoryImpl struct {
	data *data.Data

	houyiClient *microserver.HouYiConn
}

func (s *strategyRepositoryImpl) Push(ctx context.Context, strategies []*bo.Strategy) error {
	apiStrategies := build.NewBuilder().BoStrategyModelBuilder().WithBoStrategies(strategies).ToAPIs()
	strategyReply, err := s.houyiClient.PushStrategy(ctx, &strategyapi.PushStrategyRequest{
		Strategies: apiStrategies,
	})
	if err != nil {
		return err
	}
	log.Debugw("strategyReply", strategyReply)
	return nil
}
