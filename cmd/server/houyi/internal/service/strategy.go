package service

import (
	"context"

	pb "github.com/aide-family/moon/api/houyi/strategy"
)

type StrategyService struct {
	pb.UnimplementedStrategyServer
}

func NewStrategyService() *StrategyService {
	return &StrategyService{}
}

func (s *StrategyService) PushStrategy(ctx context.Context, req *pb.PushStrategyRequest) (*pb.PushStrategyReply, error) {
	return &pb.PushStrategyReply{}, nil
}
