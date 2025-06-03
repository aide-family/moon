package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewMetricManager(
	bc *conf.Bootstrap,
	metricRegisterRepo repository.MetricRegister,
	cacheRepo repository.Cache,
	logger log.Logger,
) *MetricManager {
	metricManager := &MetricManager{
		metricRegisterRepo: metricRegisterRepo,
		cacheRepo:          cacheRepo,
		helper:             log.NewHelper(log.With(logger, "module", "biz.metric")),
	}
	metricManager.loadCacheMetrics()
	metricManager.loadConfigMetrics(bc)
	return metricManager
}

type MetricManager struct {
	metricRegisterRepo repository.MetricRegister
	cacheRepo          repository.Cache
	helper             *log.Helper
}

func (m *MetricManager) WithMetricData(ctx context.Context, metrics ...*bo.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}

	metricDataList := slices.GroupBy(metrics, func(metric *bo.MetricData) vobj.MetricType {
		return metric.MetricType
	})
	safetyMetricDataList := safety.NewMap(metricDataList)

	eg := new(errgroup.Group)
	eg.Go(func() error {
		metricDataList, ok := safetyMetricDataList.Get(vobj.MetricTypeCounter)
		if !ok {
			return nil
		}
		if len(metricDataList) == 0 {
			return nil
		}
		return m.metricRegisterRepo.WithCounterMetricValue(ctx, metricDataList...)
	})
	eg.Go(func() error {
		metricDataList, ok := safetyMetricDataList.Get(vobj.MetricTypeGauge)
		if !ok {
			return nil
		}
		if len(metricDataList) == 0 {
			return nil
		}
		return m.metricRegisterRepo.WithGaugeMetricValue(ctx, metricDataList...)
	})
	eg.Go(func() error {
		metricDataList, ok := safetyMetricDataList.Get(vobj.MetricTypeHistogram)
		if !ok {
			return nil
		}
		if len(metricDataList) == 0 {
			return nil
		}
		return m.metricRegisterRepo.WithHistogramMetricValue(ctx, metricDataList...)
	})
	eg.Go(func() error {
		metricDataList, ok := safetyMetricDataList.Get(vobj.MetricTypeSummary)
		if !ok {
			return nil
		}
		if len(metricDataList) == 0 {
			return nil
		}
		return m.metricRegisterRepo.WithSummaryMetricValue(ctx, metricDataList...)
	})
	return eg.Wait()
}

func (m *MetricManager) RegisterCounterMetric(ctx context.Context, metrics ...*bo.CounterMetricVec) error {
	if len(metrics) == 0 {
		return nil
	}
	cacheMetrics := slices.Map(metrics, func(metric *bo.CounterMetricVec) bo.MetricVec {
		return metric
	})
	if err := m.cacheRepo.StorageMetric(ctx, cacheMetrics...); err != nil {
		return err
	}
	for _, metric := range metrics {
		metricValue := metric.New()
		m.metricRegisterRepo.RegisterCounterMetric(ctx, metric.GetMetricName(), metricValue)
	}
	return nil
}

func (m *MetricManager) RegisterGaugeMetric(ctx context.Context, metrics ...*bo.GaugeMetricVec) error {
	if len(metrics) == 0 {
		return nil
	}
	cacheMetrics := slices.Map(metrics, func(metric *bo.GaugeMetricVec) bo.MetricVec {
		return metric
	})
	if err := m.cacheRepo.StorageMetric(ctx, cacheMetrics...); err != nil {
		return err
	}
	for _, metric := range metrics {
		metricValue := metric.New()
		m.metricRegisterRepo.RegisterGaugeMetric(ctx, metric.GetMetricName(), metricValue)
	}
	return nil
}

func (m *MetricManager) RegisterHistogramMetric(ctx context.Context, metrics ...*bo.HistogramMetricVec) error {
	if len(metrics) == 0 {
		return nil
	}
	cacheMetrics := slices.Map(metrics, func(metric *bo.HistogramMetricVec) bo.MetricVec {
		return metric
	})
	if err := m.cacheRepo.StorageMetric(ctx, cacheMetrics...); err != nil {
		return err
	}
	for _, metric := range metrics {
		metricValue := metric.New()
		m.metricRegisterRepo.RegisterHistogramMetric(ctx, metric.GetMetricName(), metricValue)
	}
	return nil
}

func (m *MetricManager) RegisterSummaryMetric(ctx context.Context, metrics ...*bo.SummaryMetricVec) error {
	if len(metrics) == 0 {
		return nil
	}
	cacheMetrics := slices.Map(metrics, func(metric *bo.SummaryMetricVec) bo.MetricVec {
		return metric
	})
	if err := m.cacheRepo.StorageMetric(ctx, cacheMetrics...); err != nil {
		return err
	}
	for _, metric := range metrics {
		metricValue := metric.New()
		m.metricRegisterRepo.RegisterSummaryMetric(ctx, metric.GetMetricName(), metricValue)
	}
	return nil
}

func (m *MetricManager) loadCacheMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	eg := new(errgroup.Group)
	eg.Go(func() error {
		counterMetrics, err := m.cacheRepo.GetCounterMetrics(ctx)
		if err != nil {
			return err
		}
		return m.RegisterCounterMetric(ctx, counterMetrics...)
	})
	eg.Go(func() error {
		gaugeMetrics, err := m.cacheRepo.GetGaugeMetrics(ctx)
		if err != nil {
			return err
		}
		return m.RegisterGaugeMetric(ctx, gaugeMetrics...)
	})
	eg.Go(func() error {
		histogramMetrics, err := m.cacheRepo.GetHistogramMetrics(ctx)
		if err != nil {
			return err
		}
		return m.RegisterHistogramMetric(ctx, histogramMetrics...)
	})
	eg.Go(func() error {
		summaryMetrics, err := m.cacheRepo.GetSummaryMetrics(ctx)
		if err != nil {
			return err
		}
		return m.RegisterSummaryMetric(ctx, summaryMetrics...)
	})
	if err := eg.Wait(); err != nil {
		m.helper.WithContext(ctx).Errorw("method", "loadCacheMetrics", "err", err)
	}
}

func (m *MetricManager) loadConfigMetrics(bc *conf.Bootstrap) {
	metricVecs := bc.GetMetricVecs()
	if len(metricVecs) == 0 {
		return
	}
	metrics := slices.GroupBy(metricVecs, func(v *conf.MetricVec) vobj.MetricType {
		return vobj.MetricType(v.GetType())
	})
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	eg := new(errgroup.Group)
	for metricType, metrics := range metrics {
		if len(metrics) == 0 {
			continue
		}
		switch metricType {
		case vobj.MetricTypeCounter:
			eg.Go(func() error {
				counterMetrics := slices.Map(metrics, toCounterMetricVec)
				return m.RegisterCounterMetric(ctx, counterMetrics...)
			})
		case vobj.MetricTypeGauge:
			eg.Go(func() error {
				gaugeMetrics := slices.Map(metrics, toGaugeMetricVec)
				return m.RegisterGaugeMetric(ctx, gaugeMetrics...)
			})
		case vobj.MetricTypeHistogram:
			eg.Go(func() error {
				histogramMetrics := slices.Map(metrics, toHistogramMetricVec)
				return m.RegisterHistogramMetric(ctx, histogramMetrics...)
			})
		case vobj.MetricTypeSummary:
			eg.Go(func() error {
				summaryMetrics := slices.Map(metrics, toSummaryMetricVec)
				return m.RegisterSummaryMetric(ctx, summaryMetrics...)
			})
		}
	}
	if err := eg.Wait(); err != nil {
		m.helper.WithContext(ctx).Errorw("method", "loadConfigMetrics", "err", err)
	}
}

func toCounterMetricVec(metric *conf.MetricVec) *bo.CounterMetricVec {
	return &bo.CounterMetricVec{
		Namespace: metric.GetNamespace(),
		SubSystem: metric.GetSubSystem(),
		Name:      metric.GetName(),
		Labels:    metric.GetLabels(),
		Help:      metric.GetHelp(),
	}
}

func toGaugeMetricVec(metric *conf.MetricVec) *bo.GaugeMetricVec {
	return &bo.GaugeMetricVec{
		Namespace: metric.GetNamespace(),
		SubSystem: metric.GetSubSystem(),
		Name:      metric.GetName(),
		Labels:    metric.GetLabels(),
		Help:      metric.GetHelp(),
	}
}

func toHistogramMetricVec(metric *conf.MetricVec) *bo.HistogramMetricVec {
	return &bo.HistogramMetricVec{
		Namespace:                       metric.GetNamespace(),
		SubSystem:                       metric.GetSubSystem(),
		Name:                            metric.GetName(),
		Labels:                          metric.GetLabels(),
		Help:                            metric.GetHelp(),
		Buckets:                         metric.GetBuckets(),
		NativeHistogramBucketFactor:     metric.GetNativeHistogramBucketFactor(),
		NativeHistogramZeroThreshold:    metric.GetNativeHistogramZeroThreshold(),
		NativeHistogramMaxBucketNumber:  metric.GetNativeHistogramMaxBucketNumber(),
		NativeHistogramMinResetDuration: metric.GetNativeHistogramMinResetDuration(),
		NativeHistogramMaxZeroThreshold: metric.GetNativeHistogramMaxZeroThreshold(),
		NativeHistogramMaxExemplars:     metric.GetNativeHistogramMaxExemplars(),
		NativeHistogramExemplarTTL:      metric.GetNativeHistogramExemplarTTL(),
	}
}

func toSummaryMetricVec(metric *conf.MetricVec) *bo.SummaryMetricVec {
	objectiveList := metric.GetObjectives()
	objectives := make(map[float64]float64, len(objectiveList))
	for _, objective := range objectiveList {
		objectives[objective.GetQuantile()] = objective.GetValue()
	}
	return &bo.SummaryMetricVec{
		Namespace:  metric.GetNamespace(),
		SubSystem:  metric.GetSubSystem(),
		Name:       metric.GetName(),
		Labels:     metric.GetLabels(),
		Help:       metric.GetHelp(),
		Objectives: objectives,
		MaxAge:     metric.GetMaxAge(),
		AgeBuckets: metric.GetAgeBuckets(),
		BufCap:     metric.GetBufCap(),
	}
}
