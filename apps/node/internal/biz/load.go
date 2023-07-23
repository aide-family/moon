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
	"strings"
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
	newDirList := make([]string, 0, len(dirList))
	configPath := conf.GetConfigPath()
	// 去除configPath末尾的"/"
	if configPath != "" && configPath[len(configPath)-1] == '/' {
		configPath = configPath[:len(configPath)-1]
	}

	for _, dir := range dirList {
		newDir := dir
		// 判断是否为绝对路径
		if dir[0] != '/' {
			newDir = strings.Join([]string{configPath, dir}, "/")
		}
		newDirList = append(newDirList, newDir)
	}

	err := l.repo.LoadStrategy(ctx, newDirList)
	if err != nil {
		l.logger.Errorf("LoadStrategy err: %v", err)
		return nil, err
	}
	return &pb.ReloadReply{
		Response: &api.Response{
			Code:     0,
			Message:  "load strategies success",
			Metadata: nil,
			Data:     nil,
		},
		Timestamp: time.Now().Unix(),
	}, nil
}
