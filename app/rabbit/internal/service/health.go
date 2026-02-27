package service

import (
	"context"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/rabbit/internal/biz"
)

func NewHealthService(healthBiz *biz.Health) *HealthService {
	return &HealthService{
		healthBiz: healthBiz,
		uptime:    time.Now(),
	}
}

type HealthService struct {
	apiv1.UnimplementedHealthServer
	uptime    time.Time
	healthBiz *biz.Health
}

func (s *HealthService) HealthCheck(ctx context.Context, req *apiv1.HealthCheckRequest) (*apiv1.HealthCheckReply, error) {
	return &apiv1.HealthCheckReply{
		Status:   "OK",
		Message:  "rabbit is running",
		Uptime:   timestamppb.New(s.uptime),
		Duration: time.Since(s.uptime).String(),
	}, nil
}
