package service

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/env"
)

// HealthService 健康检查
type HealthService struct {
	api.UnimplementedHealthServer
}

// NewHealthService 创建健康检查服务
func NewHealthService() *HealthService {
	return &HealthService{}
}

// Check 检查
func (s *HealthService) Check(_ context.Context, _ *api.CheckRequest) (*api.CheckReply, error) {
	return &api.CheckReply{Healthy: true, Version: env.Version()}, nil
}
