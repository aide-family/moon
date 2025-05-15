package build

import (
	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToMetricQueryResults(results []*bo.MetricQueryResult) []*common.MetricQueryResult {
	return slices.Map(results, ToMetricQueryResult)
}

func ToMetricQueryResult(result *bo.MetricQueryResult) *common.MetricQueryResult {
	if validate.IsNil(result) {
		return nil
	}
	return &common.MetricQueryResult{
		Metric: result.Metric,
		Values: ToMetricQueryResultValues(result.Values),
		Value:  ToMetricQueryResultValue(result.Value),
	}
}

func ToMetricQueryResultValue(value *bo.MetricQueryValue) *common.MetricQueryResultValue {
	if validate.IsNil(value) {
		return nil
	}
	return &common.MetricQueryResultValue{
		Timestamp: value.Timestamp,
		Value:     value.Value,
	}
}

func ToMetricQueryResultValues(values []*bo.MetricQueryValue) []*common.MetricQueryResultValue {
	return slices.Map(values, ToMetricQueryResultValue)
}
