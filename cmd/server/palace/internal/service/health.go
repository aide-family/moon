package service

import (
	"context"

	pb "github.com/aide-family/moon/api"
)

type HealthService struct {
	pb.UnimplementedHealthServer
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check(_ context.Context, _ *pb.CheckRequest) (*pb.CheckReply, error) {
	return &pb.CheckReply{Healthy: true}, nil
}
