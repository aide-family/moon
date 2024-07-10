package service

import (
	"context"

	pb "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
)

type StrategyService struct {
	pb.UnimplementedStrategyServer

	strategyBiz *biz.StrategyBiz
}

func NewStrategyService(strategyBiz *biz.StrategyBiz) *StrategyService {
	return &StrategyService{
		strategyBiz: strategyBiz,
	}
}

func (s *StrategyService) PushStrategy(ctx context.Context, req *pb.PushStrategyRequest) (*pb.PushStrategyReply, error) {
	return &pb.PushStrategyReply{}, nil
}
