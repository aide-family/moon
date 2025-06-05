package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/encoding"

	"github.com/aide-family/moon/cmd/laurel/internal/biz"
	"github.com/aide-family/moon/cmd/laurel/internal/service/build"
	apicommon "github.com/aide-family/moon/pkg/api/laurel/common"
	apiv1 "github.com/aide-family/moon/pkg/api/laurel/v1"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewMetricService(metricManager *biz.MetricManager, logger log.Logger) *MetricService {
	return &MetricService{
		metricManager: metricManager,
		helper:        log.NewHelper(log.With(logger, "module", "service.metric")),
	}
}

type MetricService struct {
	apiv1.UnimplementedMetricServer
	metricManager *biz.MetricManager
	helper        *log.Helper
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

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return s.metricManager.RegisterCounterMetric(ctx, counterVecs...)
	})
	eg.Go(func() error {
		return s.metricManager.RegisterGaugeMetric(ctx, gaugeVecs...)
	})
	eg.Go(func() error {
		return s.metricManager.RegisterHistogramMetric(ctx, histogramVecs...)
	})
	eg.Go(func() error {
		return s.metricManager.RegisterSummaryMetric(ctx, summaryVecs...)
	})

	return &apiv1.EmptyReply{}, eg.Wait()
}

func (s *MetricService) PushMetricEvent(ctx context.Context, req []byte) {
	var metricEvent apicommon.MetricEvent
	codec := encoding.GetCodec("json")
	if err := codec.Unmarshal(req, &metricEvent); err != nil {
		s.helper.Warnf("unmarshal metric event error: %v", err)
		return
	}
	registerMetricVecs := []*apicommon.MetricVec{metricEvent.GetVec()}
	if _, err := s.RegisterMetric(ctx, &apiv1.RegisterMetricRequest{
		MetricVecs: registerMetricVecs,
	}); err != nil {
		s.helper.Warnf("register metric error: %v", err)
		return
	}
	pushMetricData := []*apicommon.MetricData{metricEvent.GetData()}
	if _, err := s.PushMetricData(ctx, &apiv1.PushMetricDataRequest{
		Metrics: pushMetricData,
	}); err != nil {
		s.helper.Warnf("push metric data error: %v", err)
		return
	}
}
