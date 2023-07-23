package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
	pb "prometheus-manager/api/strategy/v1/push"
	"prometheus-manager/apps/node/internal/service"
)

type (
	IPushRepo interface {
		V1Repo
	}

	PushLogic struct {
		logger *log.Helper
		repo   IPushRepo
		tr     trace.Tracer
	}
)

var _ service.IPushLogic = (*PushLogic)(nil)

func NewPushLogic(repo IPushRepo, logger log.Logger) *PushLogic {
	return &PushLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Push"))}
}

func (s *PushLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	return nil, nil
}
