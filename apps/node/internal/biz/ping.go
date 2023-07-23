package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	pb "prometheus-manager/api"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
)

type (
	IPingRepo interface {
		Check(ctx context.Context) (*conf.Env, error)
	}

	PingLogic struct {
		logger *log.Helper
		repo   IPingRepo
		tr     trace.Tracer
	}
)

var _ service.IPingLogic = (*PingLogic)(nil)

func NewPingLogic(repo IPingRepo, logger log.Logger) *PingLogic {
	return &PingLogic{
		repo:   repo,
		logger: log.NewHelper(log.With(logger, "module", "biz/Ping")),
		tr:     otel.Tracer("biz/Ping"),
	}
}

func (l *PingLogic) Check(ctx context.Context, req *pb.PingRequest) (*pb.PingReply, error) {
	ctx, span := l.tr.Start(ctx, "Check "+req.GetName())
	defer span.End()

	env, err := l.repo.Check(ctx)
	if err != nil {
		l.logger.Errorf("Check err: %v", err)
		return nil, err
	}

	return &pb.PingReply{
		Name:      env.GetName(),
		Version:   env.GetVersion(),
		Namespace: env.GetNamespace(),
		Metadata:  env.GetMetadata(),
	}, nil
}
