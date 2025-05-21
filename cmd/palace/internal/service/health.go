package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/helper/middleware"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	pb "github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
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

func (s *HealthService) Check(ctx context.Context, _ *pb.CheckRequest) (*pb.CheckReply, error) {
	return &pb.CheckReply{
		Healthy: true,
		Version: hello.GetEnv().Version(),
		Time:    timex.Format(timex.Now()),
	}, nil
}

func (s *HealthService) CreateOperateLog(ctx context.Context, req *middleware.OperateLogParams) {
	menuDo, ok := do.GetMenuDoContext(ctx)
	if !ok || validate.IsNil(menuDo) {
		return
	}
	if !menuDo.GetProcessType().IsContainsLog() {
		return
	}
	params := build.ToOperateLogParams(ctx, menuDo, req)
	if validate.IsNil(params) {
		return
	}
	s.logsBiz.CreateOperateLog(ctx, params)
}
