package biz

import (
	"context"
	"path"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/strategy"
	pb "prometheus-manager/api/strategy/v1/push"

	"prometheus-manager/pkg/util/dir"
	"prometheus-manager/pkg/util/stringer"

	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
)

type (
	StoreStrategyResult struct {
		SuccessCount int64
		FailedCount  int64
		StrategyDirs []*strategy.StrategyDir
	}

	IPushRepo interface {
		V1Repo
		StoreStrategy(ctx context.Context, strategyDirList []*strategy.StrategyDir) (*StoreStrategyResult, error)
		RemoveStrategy(ctx context.Context, files []string) error
	}

	PushLogic struct {
		logger *log.Helper
		repo   IPushRepo
	}
)

var _ service.IPushLogic = (*PushLogic)(nil)

func NewPushLogic(repo IPushRepo, logger log.Logger) *PushLogic {
	return &PushLogic{
		repo:   repo,
		logger: log.NewHelper(log.With(logger, "module", pushModuleName)),
	}
}

func (l *PushLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := otel.Tracer(pushModuleName).Start(ctx, "PushLogic.Strategies")
	defer span.End()

	datasource := conf.Get().GetStrategy().GetPromDatasources()
	strategyPathMap := getStrategyPathMap(datasource)

	// 判断路径是否在允许的范围内
	newStrategyDirs := make([]*strategy.StrategyDir, 0, len(req.GetStrategyDirs()))
	for _, strategyDir := range req.GetStrategyDirs() {
		dirInfo := dir.RemoveLastSlash(strategyDir.Dir)
		dirString := joinUniKey(req.GetNode(), dirInfo)
		if _, ok := strategyPathMap[dirString]; !ok {
			l.logger.Errorf("strategy path %s not allowed", dirString)
			continue
		}
		newStrategyDir := strategyDir
		newStrategyDir.Dir = dir.MakeDir(conf.GetConfigPath(), dirInfo)
		newStrategyDirs = append(newStrategyDirs, newStrategyDir)
	}

	if len(newStrategyDirs) == 0 {
		return nil, perrors.ErrorLogicUnauthorizedPath("strategy path not allowed")
	}

	storeResult, err := l.repo.StoreStrategy(ctx, newStrategyDirs)
	if err != nil {
		l.logger.Errorf("StoreStrategy error: %v", err)
		return nil, err
	}

	if storeResult == nil {
		return nil, perrors.ErrorLogicUnknown("store strategy result is nil")
	}

	return &pb.StrategiesReply{
		Response:  &api.Response{Message: l.repo.V1(ctx)},
		Timestamp: time.Now().Unix(),
		Result: &pb.Result{
			SuccessCount: storeResult.SuccessCount,
			FailedCount:  storeResult.FailedCount,
			StrategyDirs: storeResult.StrategyDirs,
		},
	}, nil
}

func (l *PushLogic) DeleteStrategies(ctx context.Context, req *pb.DeleteStrategiesRequest) (*pb.DeleteStrategiesReply, error) {
	ctx, span := otel.Tracer(pushModuleName).Start(ctx, "PushLogic.DeleteStrategies")
	defer span.End()

	dirs := req.GetDirs()
	datasource := conf.Get().GetStrategy().GetPromDatasources()
	strategyPathMap := getStrategyPathMap(datasource)

	var authorizedPath []string
	for _, strategyDir := range dirs {
		dirInfo := dir.RemoveLastSlash(strategyDir.Dir)
		dirString := joinUniKey(req.GetNode(), dirInfo)
		if _, ok := strategyPathMap[dirString]; !ok {
			l.logger.Errorf("strategy path %s not allowed", dirString)
			continue
		}

		dirInfo = dir.MakeDir(conf.GetConfigPath(), dirInfo)

		for _, filename := range strategyDir.GetFilenames() {
			authorizedPath = append(authorizedPath, path.Join(dirInfo, filename))
		}
	}

	if len(authorizedPath) > 0 {
		return nil, perrors.ErrorLogicUnauthorizedPath("strategy path not allowed")
	}

	if err := l.repo.RemoveStrategy(ctx, authorizedPath); err != nil {
		l.logger.WithContext(ctx).Errorf("RemoveStrategy error: %v", err)
		return nil, perrors.ErrorLogicDeletePrometheusStrategyFailed("delete prometheus strategy file failed").WithCause(err).WithMetadata(map[string]string{
			"authorizedPath": stringer.New(authorizedPath).String(),
		})
	}

	return &pb.DeleteStrategiesReply{Response: &api.Response{Message: "success"}}, nil
}
