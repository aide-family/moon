package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	pb "prometheus-manager/api/strategy/v1/load"
	"prometheus-manager/apps/node/internal/service"
)

type (
	ILoadRepo interface {
		V1Repo
	}

	LoadLogic struct {
		logger *log.Helper
		repo   ILoadRepo
		tr     trace.Tracer
	}
)

var _ service.ILoadLogic = (*LoadLogic)(nil)

func NewLoadLogic(repo ILoadRepo, logger log.Logger) *LoadLogic {
	return &LoadLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Load")), tr: otel.Tracer("biz/Code")}
}

func (s *LoadLogic) Reload(ctx context.Context, req *pb.ReloadRequest) (*pb.ReloadReply, error) {
	return nil, nil
}
