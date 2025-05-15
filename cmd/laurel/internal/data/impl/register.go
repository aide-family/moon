package impl

import (
	"context"
	"errors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/data"
	"github.com/aide-family/moon/pkg/merr"
)

func NewMetricRegister(data *data.Data) repository.MetricRegister {
	return &metricRegisterImpl{
		Data: data,
	}
}

type metricRegisterImpl struct {
	*data.Data
}

// WithCounterMetricValue implements repository.MetricRegister.
func (m *metricRegisterImpl) WithCounterMetricValue(ctx context.Context, metrics ...*bo.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}
	errList := make([]error, 0, len(metrics))
	for _, metric := range metrics {
		counterVec, ok := m.GetCounterMetric(metric.GetMetricName())
		if !ok {
			errList = append(errList, merr.ErrorNotFound("counter metric %s not found", metric.GetMetricName()))
			continue
		}
		counterVec.With(metric.Labels).Add(metric.Value)
	}
	if len(errList) > 0 {
		return errors.Join(errList...)
	}
	return nil
}

// WithGaugeMetricValue implements repository.MetricRegister.
func (m *metricRegisterImpl) WithGaugeMetricValue(ctx context.Context, metrics ...*bo.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}
	errList := make([]error, 0, len(metrics))
	for _, metric := range metrics {
		gaugeVec, ok := m.GetGaugeMetric(metric.GetMetricName())
		if !ok {
			errList = append(errList, merr.ErrorNotFound("gauge metric %s not found", metric.GetMetricName()))
			continue
		}
		gaugeVec.With(metric.Labels).Set(metric.Value)
	}
	if len(errList) > 0 {
		return errors.Join(errList...)
	}
	return nil
}

// WithHistogramMetricValue implements repository.MetricRegister.
func (m *metricRegisterImpl) WithHistogramMetricValue(ctx context.Context, metrics ...*bo.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}
	errList := make([]error, 0, len(metrics))
	for _, metric := range metrics {
		histogramVec, ok := m.GetHistogramMetric(metric.GetMetricName())
		if !ok {
			errList = append(errList, merr.ErrorNotFound("histogram metric %s not found", metric.GetMetricName()))
			continue
		}
		histogramVec.With(metric.Labels).Observe(metric.Value)
	}
	if len(errList) > 0 {
		return errors.Join(errList...)
	}
	return nil
}

// WithSummaryMetricValue implements repository.MetricRegister.
func (m *metricRegisterImpl) WithSummaryMetricValue(ctx context.Context, metrics ...*bo.MetricData) error {
	if len(metrics) == 0 {
		return nil
	}
	errList := make([]error, 0, len(metrics))
	for _, metric := range metrics {
		summaryVec, ok := m.GetSummaryMetric(metric.GetMetricName())
		if !ok {
			errList = append(errList, merr.ErrorNotFound("summary metric %s not found", metric.GetMetricName()))
			continue
		}
		summaryVec.With(metric.Labels).Observe(metric.Value)
	}
	if len(errList) > 0 {
		return errors.Join(errList...)
	}
	return nil
}

// RegisterCounterMetric implements repository.MetricRegister.
// Subtle: this method shadows the method (*Data).RegisterCounterMetric of metricRegisterImpl.Data.
func (m *metricRegisterImpl) RegisterCounterMetric(ctx context.Context, name string, metric *prometheus.CounterVec) {
	if !m.SetCounterMetric(name, metric) {
		return
	}
	prometheus.MustRegister(metric)
}

// RegisterGaugeMetric implements repository.MetricRegister.
// Subtle: this method shadows the method (*Data).RegisterGaugeMetric of metricRegisterImpl.Data.
func (m *metricRegisterImpl) RegisterGaugeMetric(ctx context.Context, name string, metric *prometheus.GaugeVec) {
	if !m.SetGaugeMetric(name, metric) {
		return
	}
	prometheus.MustRegister(metric)
}

// RegisterHistogramMetric implements repository.MetricRegister.
// Subtle: this method shadows the method (*Data).RegisterHistogramMetric of metricRegisterImpl.Data.
func (m *metricRegisterImpl) RegisterHistogramMetric(ctx context.Context, name string, metric *prometheus.HistogramVec) {
	if !m.SetHistogramMetric(name, metric) {
		return
	}
	prometheus.MustRegister(metric)
}

// RegisterSummaryMetric implements repository.MetricRegister.
// Subtle: this method shadows the method (*Data).RegisterSummaryMetric of metricRegisterImpl.Data.
func (m *metricRegisterImpl) RegisterSummaryMetric(ctx context.Context, name string, metric *prometheus.SummaryVec) {
	if !m.SetSummaryMetric(name, metric) {
		return
	}
	prometheus.MustRegister(metric)
}
