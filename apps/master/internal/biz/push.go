package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	pb "prometheus-manager/api/node"
	"prometheus-manager/apps/master/internal/service"
)

type (
	IPushRepo interface {
		V1Repo
	}

	PushLogic struct {
		logger *log.Helper
		repo   IPushRepo
	}
)

var _ service.IPushLogic = (*PushLogic)(nil)

func NewPushLogic(repo IPushRepo, logger log.Logger) *PushLogic {
	return &PushLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Push"))}
}

func (s *PushLogic) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PushLogic.Call")
	defer span.End()
	return &pb.CallResponse{Response: &api.Response{
		Code:    0,
		Message: s.repo.V1(ctx),
	}}, nil
}
