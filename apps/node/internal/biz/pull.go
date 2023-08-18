package biz

import (
	"context"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api/perrors"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/pkg/util/stringer"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/strategy"
	pb "prometheus-manager/api/strategy/v1/pull"

	"prometheus-manager/apps/node/internal/service"
)

type (
	StrategyLoad struct {
		StrategyDirs []*strategy.StrategyDir `json:"strategy"`
		LoadTime     time.Time               `json:"load_time"`
	}

	IPullRepo interface {
		V1Repo
		PullStrategies(ctx context.Context) (*StrategyLoad, error)
		Datasources(ctx context.Context) (*pb.DatasourcesReply, error)
	}

	PullLogic struct {
		logger *log.Helper
		repo   IPullRepo
	}
)

var _ service.IPullLogic = (*PullLogic)(nil)

func NewPullLogic(repo IPullRepo, logger log.Logger) *PullLogic {
	return &PullLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Pull"))}
}

func (s *PullLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := otel.Tracer("biz/pull").Start(ctx, "PullLogic.Strategies")
	defer span.End()

	strategyLoad, err := s.repo.PullStrategies(ctx)
	if err != nil {
		s.logger.Errorf("PullStrategies err: %v", err)
		return nil, err
	}

	return &pb.StrategiesReply{StrategyDirs: strategyLoad.StrategyDirs, Timestamp: strategyLoad.LoadTime.Unix()}, nil
}

func (s *PullLogic) Datasources(ctx context.Context, req *pb.DatasourcesRequest) (*pb.DatasourcesReply, error) {
	ctx, span := otel.Tracer("biz/pull").Start(ctx, "PullLogic.Datasources")
	defer span.End()

	serverEnv := conf.Get().GetEnv()
	if serverEnv == nil || req.GetNode() != serverEnv.GetName() {
		s.logger.Errorf("node not found: %s", req.GetNode())
		return nil, perrors.ErrorClientNotFound("node not found").WithMetadata(map[string]string{
			"node": req.GetNode(),
			"env":  stringer.New(serverEnv).String(),
		})
	}

	return s.repo.Datasources(ctx)
}
