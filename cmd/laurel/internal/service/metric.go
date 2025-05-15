package service

import (
	"context"

	"github.com/moon-monitor/moon/cmd/laurel/internal/biz"
	"github.com/moon-monitor/moon/cmd/laurel/internal/service/build"
	apicommon "github.com/moon-monitor/moon/pkg/api/laurel/common"
	apiv1 "github.com/moon-monitor/moon/pkg/api/laurel/v1"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

func NewMetricService(metricManager *biz.MetricManager) *MetricService {
	return &MetricService{
		metricManager: metricManager,
	}
}

type MetricService struct {
	apiv1.UnimplementedMetricServer
	metricManager *biz.MetricManager
}

func (s *MetricService) PushMetricData(ctx context.Context, req *apiv1.PushMetricDataRequest) (*apiv1.EmptyReply, error) {
	metricDataList := build.ToMetricDataList(req.GetMetrics())
	if err := s.metricManager.WithMetricData(ctx, metricDataList...); err != nil {
		return nil, err
	}
	return &apiv1.EmptyReply{}, nil
}

func (s *MetricService) RegisterMetric(ctx context.Context, req *apiv1.RegisterMetricRequest) (*apiv1.EmptyReply, error) {
	metricVecs := slices.GroupBy(req.GetMetricVecs(), func(v *apicommon.MetricVec) apicommon.MetricType {
		return v.GetMetricType()
	})
	counterVecs := build.ToCounterMetricVecs(metricVecs[apicommon.MetricType_METRIC_TYPE_COUNTER])
	gaugeVecs := build.ToGaugeMetricVecs(metricVecs[apicommon.MetricType_METRIC_TYPE_GAUGE])
	histogramVecs := build.ToHistogramMetricVecs(metricVecs[apicommon.MetricType_METRIC_TYPE_HISTOGRAM])
	summaryVecs := build.ToSummaryMetricVecs(metricVecs[apicommon.MetricType_METRIC_TYPE_SUMMARY])

	s.metricManager.RegisterCounterMetric(ctx, counterVecs...)
	s.metricManager.RegisterGaugeMetric(ctx, gaugeVecs...)
	s.metricManager.RegisterHistogramMetric(ctx, histogramVecs...)
	s.metricManager.RegisterSummaryMetric(ctx, summaryVecs...)

	return &apiv1.EmptyReply{}, nil
}
