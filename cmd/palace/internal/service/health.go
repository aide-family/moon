package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	pb "github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/util/timex"
)

type HealthService struct {
	pb.UnimplementedHealthServer
	logsBiz *biz.Logs
}

func NewHealthService(logsBiz *biz.Logs) *HealthService {
	return &HealthService{
		logsBiz: logsBiz,
	}
}

func (s *HealthService) Check(ctx context.Context, req *pb.CheckRequest) (*pb.CheckReply, error) {
	return &pb.CheckReply{
		Healthy: true,
		Version: hello.GetEnv().Version(),
		Time:    timex.Format(timex.Now()),
	}, nil
}

func (s *HealthService) CreateOperateLog(ctx context.Context, req *middleware.OperateLogParams) {
	s.logsBiz.CreateOperateLog(ctx, req)
}
