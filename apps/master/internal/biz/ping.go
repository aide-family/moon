package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api"
	"prometheus-manager/apps/master/internal/service"
)

type (
	IPingRepo interface {
		V1Repo
	}

	PingLogic struct {
		logger *log.Helper
		repo   IPingRepo
	}
)

var _ service.IPingLogic = (*PingLogic)(nil)

func NewPingLogic(repo IPingRepo, logger log.Logger) *PingLogic {
	return &PingLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Ping"))}
}

func (s *PingLogic) Check(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	return nil, nil
}
