package service

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz"
	"github.com/aide-family/moon/pkg/env"
)

// HealthService 健康检查
type HealthService struct {
	api.UnimplementedHealthServer

	heartbeatBiz *biz.HeartbeatBiz
}

// NewHealthService 创建健康检查服务
func NewHealthService(heartbeatBiz *biz.HeartbeatBiz) *HealthService {
	return &HealthService{
		heartbeatBiz: heartbeatBiz,
	}
}

// Check 检查
func (s *HealthService) Check(_ context.Context, _ *api.CheckRequest) (*api.CheckReply, error) {
	return &api.CheckReply{Healthy: true, Version: env.Version()}, nil
}

// Heartbeat 心跳包
func (s *HealthService) Heartbeat(ctx context.Context, in *api.HeartbeatRequest) error {
	return s.heartbeatBiz.Heartbeat(ctx, in)
}
