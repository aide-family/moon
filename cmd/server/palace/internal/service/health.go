package service

import (
	"context"

	"github.com/aide-family/moon/api"
)

type HealthService struct {
	api.UnimplementedHealthServer
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check(_ context.Context, _ *api.CheckRequest) (*api.CheckReply, error) {
	return &api.CheckReply{Healthy: true}, nil
}
