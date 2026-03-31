package service

import (
	"context"
	"time"

	healthv1 "github.com/aide-family/magicbox/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/aide-family/jade_tree/internal/biz"
)

func NewHealthService(healthBiz *biz.Health) *HealthService {
	return &HealthService{healthBiz: healthBiz, uptime: time.Now()}
}

type HealthService struct {
	healthv1.UnimplementedHealthServer
	uptime    time.Time
	healthBiz *biz.Health
}

func (s *HealthService) HealthCheck(ctx context.Context, req *healthv1.HealthCheckRequest) (*healthv1.HealthCheckReply, error) {
	return &healthv1.HealthCheckReply{Status: "OK", Message: "jade_tree is running", Uptime: timestamppb.New(s.uptime), Duration: time.Since(s.uptime).String()}, nil
}
