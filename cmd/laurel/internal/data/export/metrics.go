package export

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func NewNodeExportMetricRepo(logger log.Logger) repository.Metrics {
	return &nodeExportMetricRepoImpl{
		metrics: safety.NewMap[string, prometheus.Collector](),
		helper:  log.NewHelper(log.With(logger, "module", "laurel.data.export.metrics")),
	}
}

type nodeExportMetricRepoImpl struct {
	metrics *safety.Map[string, prometheus.Collector]
	helper  *log.Helper
}

// Metrics implements repository.Metrics.
func (n *nodeExportMetricRepoImpl) Metrics(ctx context.Context, target string) (map[string]prometheus.Collector, error) {
	resp, err := http.Get(target)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			n.helper.WithContext(ctx).Warnw("method", "prometheus.metadata", "err", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metrics: %v", err)
	}
	metrics := make(map[string]prometheus.Collector, len(metricFamilies))
	for name, metricFamily := range metricFamilies {
		var collector prometheus.Collector
		switch metricFamily.GetType() {
		case dto.MetricType_COUNTER:
			collector = n.createCounter(name, metricFamily)
		case dto.MetricType_GAUGE, dto.MetricType_UNTYPED:
			collector = n.createGauge(name, metricFamily)
		case dto.MetricType_HISTOGRAM:
			collector = n.createHistogram(name, metricFamily)
		case dto.MetricType_SUMMARY:
			collector = n.createSummary(name, metricFamily)
		}
		metrics[name] = collector
	}

	n.metrics.Append(metrics)

	return metrics, nil
}

func (n *nodeExportMetricRepoImpl) createCounter(name string, mf *dto.MetricFamily) *prometheus.CounterVec {
	labels := n.getLabelNames(mf)
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: mf.GetHelp(),
		}, labels,
	)

	for _, metric := range mf.GetMetric() {
		labelValues := n.getLabelValues(metric)
		counter.With(labelValues).Add(metric.GetCounter().GetValue())
	}

	return counter
}

func (n *nodeExportMetricRepoImpl) createGauge(name string, mf *dto.MetricFamily) *prometheus.GaugeVec {
	labels := n.getLabelNames(mf)
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: mf.GetHelp(),
		},
		labels,
	)

	for _, metric := range mf.GetMetric() {
		labelValues := n.getLabelValues(metric)
		if metric.GetGauge() != nil {
			gauge.With(labelValues).Set(metric.GetGauge().GetValue())
		} else if metric.GetUntyped() != nil {
			gauge.With(labelValues).Set(metric.GetUntyped().GetValue())
		}
	}

	return gauge
}

func (n *nodeExportMetricRepoImpl) createSummary(name string, mf *dto.MetricFamily) *prometheus.SummaryVec {
	labels := n.getLabelNames(mf)

	objectives := make(map[float64]float64)
	for _, metric := range mf.GetMetric() {
		if summaryMetric := metric.GetSummary(); summaryMetric != nil {
			for _, q := range summaryMetric.GetQuantile() {
				objectives[q.GetQuantile()] = q.GetValue()
			}
		}
	}

	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       name,
			Help:       mf.GetHelp(),
			Objectives: objectives,
		},
		labels,
	)

	for _, metric := range mf.GetMetric() {
		if summaryMetric := metric.GetSummary(); summaryMetric != nil {
			labelValues := n.getLabelValues(metric)
			summary.With(labelValues).Observe(summaryMetric.GetSampleSum() / float64(summaryMetric.GetSampleCount()))
		}
	}

	return summary
}

func (n *nodeExportMetricRepoImpl) createHistogram(name string, mf *dto.MetricFamily) *prometheus.HistogramVec {
	labels := n.getLabelNames(mf)

	var buckets []float64
	for _, metric := range mf.GetMetric() {
		if histogramMetric := metric.GetHistogram(); histogramMetric != nil {
			buckets = make([]float64, len(histogramMetric.GetBucket()))
			for i, b := range histogramMetric.GetBucket() {
				buckets[i] = b.GetUpperBound()
			}
			break
		}
	}

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    mf.GetHelp(),
			Buckets: buckets,
		},
		labels,
	)

	for _, metric := range mf.GetMetric() {
		if histogramMetric := metric.GetHistogram(); histogramMetric != nil {
			labelValues := n.getLabelValues(metric)
			histogram.With(labelValues).Observe(histogramMetric.GetSampleSum() / float64(histogramMetric.GetSampleCount()))
		}
	}

	return histogram
}

func (n *nodeExportMetricRepoImpl) getLabelNames(mf *dto.MetricFamily) []string {
	if len(mf.GetMetric()) == 0 {
		return nil
	}
	labelNames := make([]string, 0, len(mf.GetMetric()[0].GetLabel()))
	for _, label := range mf.GetMetric()[0].GetLabel() {
		labelNames = append(labelNames, label.GetName())
	}
	return labelNames
}

func (n *nodeExportMetricRepoImpl) getLabelValues(metric *dto.Metric) map[string]string {
	values := make(map[string]string, len(metric.GetLabel()))
	for _, label := range metric.GetLabel() {
		values[label.GetName()] = label.GetValue()
	}
	return values
}
