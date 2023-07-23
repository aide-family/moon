package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/strategy/v1/load"
)

type (
	ILoadLogic interface {
		Reload(ctx context.Context, req *pb.ReloadRequest) (*pb.ReloadReply, error)
	}

	LoadService struct {
		pb.UnimplementedLoadServer

		logger *log.Helper
		logic  ILoadLogic
	}
)

var _ pb.LoadServer = (*LoadService)(nil)

func NewLoadService(logic ILoadLogic, logger log.Logger) *LoadService {
	return &LoadService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Load"))}
}

func (l *LoadService) Reload(ctx context.Context, req *pb.ReloadRequest) (*pb.ReloadReply, error) {
	l.logger.Debugf("Reload req: %v", req)
	return l.logic.Reload(ctx, req)
}
