package service

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/env"
)

// HealthService 健康检查
type HealthService struct {
	api.UnimplementedHealthServer

	houyiSrv *data.HouYiConn
}

// NewHealthService 创建健康检查服务
func NewHealthService(houyiSrv *data.HouYiConn) *HealthService {
	return &HealthService{
		houyiSrv: houyiSrv,
	}
}

// Check 检查
func (s *HealthService) Check(ctx context.Context, req *api.CheckRequest) (*api.CheckReply, error) {
	//if _, err := s.houyiSrv.Health(ctx, req); err != nil {
	//	log.Warnw("houyiSrv", err)
	//}
	return &api.CheckReply{Healthy: true, Version: env.Version()}, nil
}
