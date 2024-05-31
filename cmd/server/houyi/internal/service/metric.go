package service

import (
	"context"

	"github.com/aide-cloud/moon/api"
	pb "github.com/aide-cloud/moon/api/houyi/metadata"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/service/build"
	"github.com/aide-cloud/moon/pkg/datasource/metric"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type MetricService struct {
	pb.UnimplementedMetricServer

	metricBiz *biz.MetricBiz
}

func NewMetricService(metricBiz *biz.MetricBiz) *MetricService {
	return &MetricService{
		metricBiz: metricBiz,
	}
}

func (s *MetricService) SyncMetadata(ctx context.Context, req *pb.SyncMetadataRequest) (*pb.SyncMetadataReply, error) {
	params := &bo.GetMetricsParams{
		Endpoint:    req.GetEndpoint(),
		Config:      req.GetConfig(),
		StorageType: vobj.StorageType(req.GetStorageType()),
	}
	metrics, err := s.metricBiz.SyncMetrics(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.SyncMetadataReply{
		Metrics: types.SliceTo(metrics, func(item *bo.MetricDetail) *api.MetricDetail {
			return build.NewMetricBuilder(item).ToApi()
		}),
	}, nil
}

// Query query metric data
func (s *MetricService) Query(ctx context.Context, req *pb.QueryRequest) (*pb.QueryReply, error) {
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
	return &pb.QueryReply{
		List: types.SliceTo(data, func(item *metric.QueryResponse) *api.MetricQueryResult {
			return build.NewMetricQueryBuilder(item).ToApi()
		}),
	}, nil
}
