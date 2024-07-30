package service

import (
	"context"

	"github.com/aide-family/moon/api"
	metadataapi "github.com/aide-family/moon/api/houyi/metadata"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
)

// MetricService 指标服务
type MetricService struct {
	metadataapi.UnimplementedMetricServer

	metricBiz *biz.MetricBiz
}

// NewMetricService 创建指标服务
func NewMetricService(metricBiz *biz.MetricBiz) *MetricService {
	return &MetricService{
		metricBiz: metricBiz,
	}
}

// SyncMetadata sync metric data
func (s *MetricService) SyncMetadata(ctx context.Context, req *metadataapi.SyncMetadataRequest) (*metadataapi.SyncMetadataReply, error) {
	params := &bo.GetMetricsParams{
		Endpoint:    req.GetEndpoint(),
		Config:      req.GetConfig(),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	metrics, err := s.metricBiz.SyncMetrics(ctx, params)
	if err != nil {
		return nil, err
	}
	return &metadataapi.SyncMetadataReply{
		Metrics: types.SliceTo(metrics, func(item *bo.MetricDetail) *api.MetricDetail {
			return build.NewMetricBuilder(item).ToAPI()
		}),
	}, nil
}

// Query query metric data
func (s *MetricService) Query(ctx context.Context, req *metadataapi.QueryRequest) (*metadataapi.QueryReply, error) {
	params := &bo.QueryQLParams{
		GetMetricsParams: bo.GetMetricsParams{
			Endpoint:    req.GetEndpoint(),
			Config:      req.GetConfig(),
			StorageType: vobj.StorageType(req.GetStorageType()),
		},
		QueryQL:   req.GetQuery(),
		TimeRange: req.GetRange(),
		Step:      req.GetStep(),
	}
	data, err := s.metricBiz.Query(ctx, params)
	if err != nil {
		return nil, err
	}
	return &metadataapi.QueryReply{
		List: types.SliceTo(data, func(item *metric.QueryResponse) *api.MetricQueryResult {
			return build.NewMetricQueryBuilder(item).ToAPI()
		}),
	}, nil
}

// SyncMetadataV2 sync metric data
func (s *MetricService) SyncMetadataV2(ctx context.Context, req *metadataapi.SyncMetadataV2Request) (*metadataapi.SyncMetadataV2Reply, error) {
	params := &bo.GetMetricsParams{
		Endpoint:    req.GetEndpoint(),
		Config:      req.GetConfig(),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	metrics, err := s.metricBiz.SyncMetrics(ctx, params)
	if err != nil {
		return nil, err
	}
	// 异步推送
	go func() {
		defer after.RecoverX()
		metricsLen := len(metrics)
		labelsLen := 0
		labelValuesLen := 0
		for index, item := range metrics {
			labelsLen += len(item.Labels)
			for _, labelValues := range item.Labels {
				labelValuesLen += len(labelValues)
			}
			if err := s.metricBiz.PushMetric(&bo.PushMetricParams{
				MetricDetail: item,
				DatasourceID: req.GetDatasourceId(),
				Done:         index == metricsLen-1,
				TeamID:       req.GetTeamId(),
			}); err != nil {
				log.Errorw("method", "push metric error", "err", err)
				continue
			}
		}
		log.Infow("method", "sync metric", "total", metricsLen, "labels", labelsLen, "labelValues", labelValuesLen)
	}()
	return &metadataapi.SyncMetadataV2Reply{}, nil
}
