package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"

	"github.com/aide-family/moon/cmd/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/service/build"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	"github.com/aide-family/moon/pkg/util/slices"
)

type SyncService struct {
	houyiv1.UnimplementedSyncServer

	configBiz *biz.Config
	metricBiz *biz.Metric
	helper    *log.Helper
}

func NewSyncService(
	configBiz *biz.Config,
	metricBiz *biz.Metric,
	logger log.Logger,
) *SyncService {
	return &SyncService{
		configBiz: configBiz,
		metricBiz: metricBiz,
		helper:    log.NewHelper(log.With(logger, "module", "service.sync")),
	}
}

func (s *SyncService) MetricStrategy(ctx context.Context, req *houyiv1.MetricStrategyRequest) (*houyiv1.SyncReply, error) {
	metricRules := build.ToMetricRules(req.GetStrategies())
	if len(metricRules) == 0 {
		return &houyiv1.SyncReply{}, nil
	}

	if err := s.metricBiz.SaveMetricRules(ctx, metricRules...); err != nil {
		return nil, err
	}
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) CertificateStrategy(ctx context.Context, req *houyiv1.CertificateStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) ServerPortStrategy(ctx context.Context, req *houyiv1.ServerPortStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) HttpStrategy(ctx context.Context, req *houyiv1.HttpStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) PingStrategy(ctx context.Context, req *houyiv1.PingStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) EventStrategy(ctx context.Context, req *houyiv1.EventStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) LogsStrategy(ctx context.Context, req *houyiv1.LogsStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) RemoveStrategy(ctx context.Context, req *houyiv1.RemoveStrategyRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) MetricDatasource(ctx context.Context, req *houyiv1.MetricDatasourceRequest) (*houyiv1.SyncReply, error) {
	metricDatasourceItems := slices.MapFilter(req.GetItems(), func(datasourceItem *common.MetricDatasourceItem) (bo.MetricDatasourceConfig, bool) {
		datasourceConfig, err := build.ToMetricDatasourceConfig(datasourceItem)
		if err != nil {
			s.helper.WithContext(ctx).Warnw("method", "ToMetricDatasourceConfig", "params", datasourceItem, "error", err)
			return nil, false
		}
		return datasourceConfig, true
	})
	if err := s.configBiz.SetMetricDatasourceConfig(ctx, metricDatasourceItems...); err != nil {
		return nil, err
	}
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) LogsDatasource(ctx context.Context, req *houyiv1.LogsDatasourceRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) EventDatasource(ctx context.Context, req *houyiv1.EventDatasourceRequest) (*houyiv1.SyncReply, error) {
	return &houyiv1.SyncReply{}, nil
}

func (s *SyncService) MetricMetadata(ctx context.Context, req *houyiv1.MetricMetadataRequest) (*houyiv1.SyncReply, error) {
	datasourceConfig, err := build.ToMetricDatasourceConfig(req.GetItem())
	if err != nil {
		return nil, err
	}
	params := &bo.SyncMetricMetadataRequest{
		Item:       datasourceConfig,
		OperatorId: req.GetOperatorId(),
	}
	if err := s.metricBiz.SyncMetricMetadata(ctx, params); err != nil {
		return nil, err
	}
	return &houyiv1.SyncReply{
		Code:    int32(codes.OK),
		Message: "同步中，请稍后查看",
	}, nil
}
