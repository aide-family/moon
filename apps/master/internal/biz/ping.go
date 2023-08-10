package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api"

	"prometheus-manager/apps/master/internal/conf"
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
	ctx, span := otel.Tracer("biz").Start(ctx, "PingLogic.Check")
	defer span.End()

	s.logger.WithContext(ctx).Infof("PingLogic.Check with req: %v", req)
	return &pb.PingReply{
		Name:      conf.Get().GetEnv().GetName(),
		Version:   s.repo.V1(ctx),
		Namespace: conf.Get().GetEnv().GetNamespace(),
		Metadata:  conf.Get().GetEnv().GetMetadata(),
	}, nil
}
