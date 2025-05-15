package build

import (
	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	apicommon "github.com/aide-family/moon/pkg/api/laurel/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToCounterMetricVecs(counterVecs []*apicommon.MetricVec) []*bo.CounterMetricVec {
	if len(counterVecs) == 0 {
		return nil
	}
	return slices.Map(counterVecs, ToCounterMetricVec)
}

func ToGaugeMetricVecs(gaugeVecs []*apicommon.MetricVec) []*bo.GaugeMetricVec {
	if len(gaugeVecs) == 0 {
		return nil
	}
	return slices.Map(gaugeVecs, ToGaugeMetricVec)
}

func ToHistogramMetricVecs(histogramVecs []*apicommon.MetricVec) []*bo.HistogramMetricVec {
	if len(histogramVecs) == 0 {
		return nil
	}
	return slices.Map(histogramVecs, ToHistogramMetricVec)
}

func ToSummaryMetricVecs(summaryVecs []*apicommon.MetricVec) []*bo.SummaryMetricVec {
	if len(summaryVecs) == 0 {
		return nil
	}
	return slices.Map(summaryVecs, ToSummaryMetricVec)
}

func ToCounterMetricVec(counterVec *apicommon.MetricVec) *bo.CounterMetricVec {
	if validate.IsNil(counterVec) {
		return nil
	}
	return &bo.CounterMetricVec{
		Namespace: counterVec.GetNamespace(),
		SubSystem: counterVec.GetSubSystem(),
		Name:      counterVec.GetName(),
		Labels:    counterVec.GetLabels(),
		Help:      counterVec.GetHelp(),
	}
}

func ToGaugeMetricVec(gaugeVec *apicommon.MetricVec) *bo.GaugeMetricVec {
	if validate.IsNil(gaugeVec) {
		return nil
	}
	return &bo.GaugeMetricVec{
		Namespace: gaugeVec.GetNamespace(),
		SubSystem: gaugeVec.GetSubSystem(),
		Name:      gaugeVec.GetName(),
		Labels:    gaugeVec.GetLabels(),
		Help:      gaugeVec.GetHelp(),
	}
}

func ToHistogramMetricVec(histogramVec *apicommon.MetricVec) *bo.HistogramMetricVec {
	if validate.IsNil(histogramVec) {
		return nil
	}
	return &bo.HistogramMetricVec{
		Namespace:                       histogramVec.GetNamespace(),
		SubSystem:                       histogramVec.GetSubSystem(),
		Name:                            histogramVec.GetName(),
		Labels:                          histogramVec.GetLabels(),
		Help:                            histogramVec.GetHelp(),
		Buckets:                         histogramVec.GetNativeHistogramBuckets(),
		NativeHistogramBucketFactor:     histogramVec.GetNativeHistogramBucketFactor(),
		NativeHistogramZeroThreshold:    histogramVec.GetNativeHistogramZeroThreshold(),
		NativeHistogramMaxBucketNumber:  histogramVec.GetNativeHistogramMaxBucketNumber(),
		NativeHistogramMinResetDuration: histogramVec.GetNativeHistogramMinResetDuration(),
		NativeHistogramMaxZeroThreshold: histogramVec.GetNativeHistogramMaxZeroThreshold(),
		NativeHistogramMaxExemplars:     histogramVec.GetNativeHistogramMaxExemplars(),
		NativeHistogramExemplarTTL:      histogramVec.GetNativeHistogramExemplarTTL(),
	}
}

func ToSummaryMetricVec(summaryVec *apicommon.MetricVec) *bo.SummaryMetricVec {
	if validate.IsNil(summaryVec) {
		return nil
	}
	objectivesList := summaryVec.GetSummaryObjectives()
	objectives := make(map[float64]float64, len(objectivesList))
	for _, objective := range objectivesList {
		objectives[objective.GetQuantile()] = objective.GetValue()
	}
	return &bo.SummaryMetricVec{
		Namespace:  summaryVec.GetNamespace(),
		SubSystem:  summaryVec.GetSubSystem(),
		Name:       summaryVec.GetName(),
		Labels:     summaryVec.GetLabels(),
		Help:       summaryVec.GetHelp(),
		Objectives: objectives,
		MaxAge:     summaryVec.GetSummaryMaxAge(),
		AgeBuckets: summaryVec.GetSummaryAgeBuckets(),
		BufCap:     summaryVec.GetSummaryBufCap(),
	}
}

func ToMetricDataList(metrics []*apicommon.MetricData) []*bo.MetricData {
	if len(metrics) == 0 {
		return nil
	}
	return slices.Map(metrics, ToMetricData)
}

func ToMetricData(metric *apicommon.MetricData) *bo.MetricData {
	if validate.IsNil(metric) {
		return nil
	}
	return &bo.MetricData{
		MetricType: vobj.MetricType(metric.GetMetricType()),
		Namespace:  metric.GetNamespace(),
		SubSystem:  metric.GetSubSystem(),
		Name:       metric.GetName(),
		Labels:     metric.GetLabels(),
		Value:      metric.GetValue(),
	}
}
