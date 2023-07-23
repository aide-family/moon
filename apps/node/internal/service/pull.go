package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/strategy/v1/pull"
)

type (
	IPullLogic interface {
		Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error)
	}

	PullService struct {
		pb.UnimplementedPullServer

		logger *log.Helper
		logic  IPullLogic
	}
)

var _ pb.PullServer = (*PullService)(nil)

func NewPullService(logic IPullLogic, logger log.Logger) *PullService {
	return &PullService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Pull"))}
}

func (l *PullService) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	l.logger.Debugf("Strategies req: %v", req)
	return l.logic.Strategies(ctx, req)
}
