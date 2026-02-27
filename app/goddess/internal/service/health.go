package service

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aide-family/goddess/internal/biz"
	magicboxv1 "github.com/aide-family/magicbox/api/v1"
)

func NewHealthService(healthBiz *biz.Health) *HealthService {
	return &HealthService{
		healthBiz: healthBiz,
	}
}

type HealthService struct {
	magicboxv1.UnimplementedHealthServer

	healthBiz *biz.Health
}

func (s *HealthService) HealthCheck(ctx context.Context, req *magicboxv1.HealthCheckRequest) (*magicboxv1.HealthCheckReply, error) {
	return &magicboxv1.HealthCheckReply{
		Status:  "OK",
		Message: "Goddess is running",
		Uptime:  timestamppb.Now(),
	}, nil
}
