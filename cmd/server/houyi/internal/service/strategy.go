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

// StrategyService 策略服务
type StrategyService struct {
	strategyapi.UnimplementedStrategyServer

	strategyBiz *biz.StrategyBiz
}

// NewStrategyService 创建策略服务
func NewStrategyService(strategyBiz *biz.StrategyBiz) *StrategyService {
	return &StrategyService{
		strategyBiz: strategyBiz,
	}
}

// PushStrategy 推送策略
func (s *StrategyService) PushStrategy(ctx context.Context, req *strategyapi.PushStrategyRequest) (*strategyapi.PushStrategyReply, error) {
	strategies := make([]bo.IStrategy, 0, len(req.GetStrategies())+len(req.GetDomainStrategies())+len(req.GetHttpStrategies())+len(req.GetPingStrategies()))
	if len(req.GetStrategies()) > 0 {
		strategies = append(strategies, types.SliceTo(req.GetStrategies(), func(item *api.MetricStrategyItem) bo.IStrategy {
			return build.NewStrategyBuilder(item).ToBo()
		})...)
	}
	if len(req.GetDomainStrategies()) > 0 {
		strategies = append(strategies, types.SliceTo(req.GetDomainStrategies(), func(item *api.DomainStrategyItem) bo.IStrategy {
			return build.NewDomainStrategyBuilder(item).ToBo()
		})...)
	}
	if len(req.GetHttpStrategies()) > 0 {
		strategies = append(strategies, types.SliceTo(req.GetHttpStrategies(), func(item *api.HttpStrategyItem) bo.IStrategy {
			return build.NewHTTPStrategyBuilder(item).ToBo()
		})...)
	}
	if len(req.GetPingStrategies()) > 0 {
		strategies = append(strategies, types.SliceTo(req.GetPingStrategies(), func(item *api.PingStrategyItem) bo.IStrategy {
			return build.NewPingStrategyBuilder(item).ToBo()
		})...)
	}
	if len(strategies) == 0 {
		return &strategyapi.PushStrategyReply{}, nil
	}

	if err := s.strategyBiz.SaveStrategy(ctx, strategies); err != nil {
		return nil, err
	}
	return &strategyapi.PushStrategyReply{}, nil
}
