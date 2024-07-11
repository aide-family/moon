package service

import (
	"context"

	"github.com/aide-family/moon/api"
	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
	"github.com/aide-family/moon/pkg/util/types"
)

type StrategyService struct {
	strategyapi.UnimplementedStrategyServer

	strategyBiz *biz.StrategyBiz
}

func NewStrategyService(strategyBiz *biz.StrategyBiz) *StrategyService {
	return &StrategyService{
		strategyBiz: strategyBiz,
	}
}

func (s *StrategyService) PushStrategy(ctx context.Context, req *strategyapi.PushStrategyRequest) (*strategyapi.PushStrategyReply, error) {
	strategies := types.SliceTo(req.GetStrategies(), func(item *api.Strategy) *bo.Strategy {
		return build.NewStrategyApiBuilder(item).ToBo()
	})
	if err := s.strategyBiz.SaveStrategy(ctx, strategies); err != nil {
		return nil, err
	}
	return &strategyapi.PushStrategyReply{}, nil
}
