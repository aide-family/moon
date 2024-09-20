package service

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-family/moon/pkg/env"
)

// HealthService 健康检查
type HealthService struct {
	api.UnimplementedHealthServer

	houyiSrv *microserver.HouYiConn
}

// NewHealthService 创建健康检查服务
func NewHealthService(houyiSrv *microserver.HouYiConn) *HealthService {
	return &HealthService{
		houyiSrv: houyiSrv,
	}
}

// Check 检查
func (s *HealthService) Check(ctx context.Context, _ *api.CheckRequest) (*api.CheckReply, error) {
	s.houyiSrv.Health(ctx)
	return &api.CheckReply{Healthy: true, Version: env.Version()}, nil
}
