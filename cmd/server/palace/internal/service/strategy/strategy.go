package strategy

import (
	"context"

	strategyapi "github.com/aide-family/moon/api/admin/strategy"
)

type Service struct {
	strategyapi.UnimplementedStrategyServer
}

func NewStrategyService() *Service {
	return &Service{}
}

func (s *Service) CreateStrategyGroup(ctx context.Context, req *strategyapi.CreateStrategyGroupRequest) (*strategyapi.CreateStrategyGroupReply, error) {
	return &strategyapi.CreateStrategyGroupReply{}, nil
}

func (s *Service) DeleteStrategyGroup(ctx context.Context, req *strategyapi.DeleteStrategyGroupRequest) (*strategyapi.DeleteStrategyGroupReply, error) {
	return &strategyapi.DeleteStrategyGroupReply{}, nil
}

func (s *Service) ListStrategyGroup(ctx context.Context, req *strategyapi.ListStrategyGroupRequest) (*strategyapi.ListStrategyGroupReply, error) {
	return &strategyapi.ListStrategyGroupReply{}, nil
}

func (s *Service) GetStrategyGroup(ctx context.Context, req *strategyapi.GetStrategyGroupRequest) (*strategyapi.GetStrategyGroupReply, error) {
	return &strategyapi.GetStrategyGroupReply{}, nil
}

func (s *Service) UpdateStrategyGroup(ctx context.Context, req *strategyapi.UpdateStrategyGroupRequest) (*strategyapi.UpdateStrategyGroupReply, error) {
	return &strategyapi.UpdateStrategyGroupReply{}, nil
}

func (s *Service) CreateStrategy(ctx context.Context, req *strategyapi.CreateStrategyRequest) (*strategyapi.CreateStrategyReply, error) {
	return &strategyapi.CreateStrategyReply{}, nil
}

func (s *Service) UpdateStrategy(ctx context.Context, req *strategyapi.UpdateStrategyRequest) (*strategyapi.UpdateStrategyReply, error) {
	return &strategyapi.UpdateStrategyReply{}, nil
}

func (s *Service) DeleteStrategy(ctx context.Context, req *strategyapi.DeleteStrategyRequest) (*strategyapi.DeleteStrategyReply, error) {
	return &strategyapi.DeleteStrategyReply{}, nil
}

func (s *Service) GetStrategy(ctx context.Context, req *strategyapi.GetStrategyRequest) (*strategyapi.GetStrategyReply, error) {
	return &strategyapi.GetStrategyReply{}, nil
}

func (s *Service) ListStrategy(ctx context.Context, req *strategyapi.ListStrategyRequest) (*strategyapi.ListStrategyReply, error) {
	return &strategyapi.ListStrategyReply{}, nil
}
