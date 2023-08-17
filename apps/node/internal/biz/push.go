package biz

import (
	"context"
	"path"
	"prometheus-manager/pkg/util/stringer"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/strategy"
	pb "prometheus-manager/api/strategy/v1/push"

	"prometheus-manager/pkg/util/dir"

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
		tr     trace.Tracer
	}
)

var _ service.IPushLogic = (*PushLogic)(nil)

func NewPushLogic(repo IPushRepo, logger log.Logger) *PushLogic {
	return &PushLogic{
		repo:   repo,
		logger: log.NewHelper(log.With(logger, "module", "biz/Push")),
		tr:     otel.Tracer("biz/Push"),
	}
}

func (l *PushLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := l.tr.Start(ctx, "Strategies")
	defer span.End()

	strategyPath := conf.Get().GetStrategy().GetPath()
	if len(strategyPath) == 0 {
		l.logger.Error("strategy path not configured")
		return nil, perrors.ErrorLogicStrategyPathNotConfigured("strategy path not configured")
	}

	strategyPathMap := make(map[string]struct{})
	for _, p := range strategyPath {
		strategyPathMap[dir.RemoveLastSlash(p)] = struct{}{}
	}

	// 判断路径是否在允许的范围内
	newStrategyDirs := req.GetStrategyDirs()
	// 未授权路径列表
	var unauthorizedPath []string
	for index, strategyDir := range newStrategyDirs {
		dirString := dir.RemoveLastSlash(strategyDir.Dir)
		if _, ok := strategyPathMap[dirString]; !ok {
			l.logger.Errorf("strategy path %s not allowed", dirString)
			unauthorizedPath = append(unauthorizedPath, dirString)
		}
		newStrategyDirs[index].Dir = dir.MakeDir(conf.GetConfigPath(), strategyDir.Dir)
	}

	if len(unauthorizedPath) > 0 {
		return nil, perrors.ErrorLogicUnauthorizedPath("strategy path [%s] not allowed", strings.Join(unauthorizedPath, ","))
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
	ctx, span := l.tr.Start(ctx, "DeleteStrategies")
	defer span.End()

	dirs := req.GetDirs()

	strategyPath := conf.Get().GetStrategy().GetPath()
	if len(strategyPath) == 0 {
		l.logger.Error("strategy path not configured")
		return nil, perrors.ErrorLogicStrategyPathNotConfigured("strategy path not configured")
	}

	strategyPathMap := make(map[string]struct{})
	for _, p := range strategyPath {
		strategyPathMap[dir.RemoveLastSlash(p)] = struct{}{}
	}

	var unauthorizedPath []string
	var authorizedPath []string
	for _, strategyDir := range dirs {
		dirString := dir.RemoveLastSlash(strategyDir.GetDir())
		if _, ok := strategyPathMap[dirString]; !ok {
			l.logger.Errorf("strategy path %s not allowed", dirString)
			unauthorizedPath = append(unauthorizedPath, dirString)
			continue
		}

		dirString = dir.MakeDir(conf.GetConfigPath(), dirString)

		for _, filename := range strategyDir.GetFilenames() {
			authorizedPath = append(authorizedPath, path.Join(dirString, filename))
		}
	}

	if len(unauthorizedPath) > 0 {
		return nil, perrors.ErrorLogicUnauthorizedPath("strategy path [%s] not allowed", strings.Join(unauthorizedPath, ","))
	}

	if err := l.repo.RemoveStrategy(ctx, authorizedPath); err != nil {
		l.logger.WithContext(ctx).Errorf("RemoveStrategy error: %v", err)
		return nil, perrors.ErrorLogicDeletePrometheusStrategyFailed("delete prometheus strategy file failed").WithCause(err).WithMetadata(map[string]string{
			"authorizedPath": stringer.New(authorizedPath).String(),
		})
	}

	return &pb.DeleteStrategiesReply{Response: &api.Response{Message: "success"}}, nil
}
