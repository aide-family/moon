package service

import (
	"context"

	pb "github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/util/timex"
)

type HealthService struct {
	pb.UnimplementedHealthServer
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckReply, error) {
	return &pb.CheckReply{
		Healthy: true,
		Version: hello.GetEnv().Version(),
		Time:    timex.Format(timex.Now()),
	}, nil
}
