package service

import (
	"context"

	"github.com/aide-cloud/moon/api"
	pb "github.com/aide-cloud/moon/api/houyi/metadata"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/service/build"
	"github.com/aide-cloud/moon/pkg/types"
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

func (s *MetricService) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncReply, error) {
	params := &bo.GetMetricsParams{
		Endpoint: req.GetEndpoint(),
		Config:   req.GetConfig(),
	}
	metrics, err := s.metricBiz.SyncMetrics(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pb.SyncReply{
		Metrics: types.SliceTo(metrics, func(item *bo.MetricDetail) *api.MetricDetail {
			return build.NewMetricBuilder(item).ToApi()
		}),
	}, nil
}
