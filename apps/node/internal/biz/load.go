package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"prometheus-manager/api"
	pb "prometheus-manager/api/strategy/v1/load"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
	"prometheus-manager/pkg/curl"
	"prometheus-manager/pkg/util/dir"
	"time"
)

type (
	ILoadRepo interface {
		V1Repo
		LoadStrategy(ctx context.Context, path []string) error
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

func (l *LoadLogic) Reload(ctx context.Context, _ *pb.ReloadRequest) (*pb.ReloadReply, error) {
	ctx, span := l.tr.Start(ctx, "Reload")
	defer span.End()

	dirList := conf.Get().GetStrategy().GetPath()
	configPath := conf.GetConfigPath()
	err := l.repo.LoadStrategy(ctx, dir.MakeDirs(configPath, dirList...))
	if err != nil {
		l.logger.Errorf("LoadStrategy err: %v", err)
		return nil, err
	}

	out, err := curl.Curl(ctx, conf.Get().GetStrategy().GetReloadPath())
	if err != nil {
		l.logger.Errorf("Curl err: %v", err)
		return nil, err
	}

	l.logger.Infof("Curl out: %v", out)

	return &pb.ReloadReply{
		Response: &api.Response{
			Code:     0,
			Message:  l.repo.V1(ctx),
			Metadata: nil,
			Data:     nil,
		},
		Timestamp: time.Now().Unix(),
	}, nil
}
