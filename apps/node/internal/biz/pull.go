package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

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
	}

	PullLogic struct {
		logger *log.Helper
		repo   IPullRepo
		tr     trace.Tracer
	}
)

var _ service.IPullLogic = (*PullLogic)(nil)

func NewPullLogic(repo IPullRepo, logger log.Logger) *PullLogic {
	return &PullLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Pull")), tr: otel.Tracer("biz/Pull")}
}

func (s *PullLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := s.tr.Start(ctx, "Strategies")
	defer span.End()

	strategyLoad, err := s.repo.PullStrategies(ctx)
	if err != nil {
		s.logger.Errorf("PullStrategies err: %v", err)
		return nil, err
	}

	return &pb.StrategiesReply{StrategyDirs: strategyLoad.StrategyDirs, Timestamp: strategyLoad.LoadTime.Unix()}, nil
}
