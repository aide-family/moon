package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/health"
)

type HealthService struct {
	v1.UnimplementedHealthServiceServer
	prober *health.Prober
}

func NewHealthService(prober *health.Prober) *HealthService {
	return &HealthService{prober: prober}
}

func (s *HealthService) Check(context.Context, *v1.CheckRequest) (*v1.CheckResponse, error) {
	return &v1.CheckResponse{Status: s.prober.Status()}, nil
}

// Ready 供 /readyz 使用：检查 PostgreSQL 与 Redis 是否可达。
func (s *HealthService) Ready(ctx context.Context) error {
	return s.prober.Ready(ctx)
}
