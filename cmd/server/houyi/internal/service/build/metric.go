package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
	"github.com/aide-family/moon/pkg/util/types"
)

func NewMetricBuilder(metricDetail *bo.MetricDetail) *MetricBuilder {
	return &MetricBuilder{
		MetricDetail: metricDetail,
	}
}

type MetricBuilder struct {
	*bo.MetricDetail
}

// ToApi 转换为api对象
func (b *MetricBuilder) ToApi() *api.MetricDetail {
	if types.IsNil(b) || types.IsNil(b.MetricDetail) {
		return nil
	}
	labels := make(map[string]*api.MetricLabelValues, len(b.Labels))
	for label, values := range b.Labels {
		labels[label] = &api.MetricLabelValues{
			Values: values,
		}
	}
	return &api.MetricDetail{
		Name:   b.Name,
		Help:   b.Help,
		Type:   b.Type,
		Labels: labels,
		Unit:   b.Unit,
	}
}

func NewMetricQueryBuilder(queryResponse *metric.QueryResponse) *MetricQueryBuilder {
	return &MetricQueryBuilder{
		QueryResponse: queryResponse,
	}
}

type MetricQueryBuilder struct {
	*metric.QueryResponse
}

// ToApi 转换为api对象
func (b *MetricQueryBuilder) ToApi() *api.MetricQueryResult {
	if types.IsNil(b) || types.IsNil(b.QueryResponse) {
		return nil
	}
	var value *api.MetricQueryValue
	var values []*api.MetricQueryValue
	if !types.IsNil(b.Value) {
		value = &api.MetricQueryValue{
			Value:     b.Value.Value,
			Timestamp: b.Value.Timestamp,
		}
	}
	if !types.IsNil(b.Values) {
		values = types.SliceToWithFilter(b.Values, func(item *metric.QueryValue) (*api.MetricQueryValue, bool) {
			if types.IsNil(item) {
				return nil, false
			}
			return &api.MetricQueryValue{
				Value:     item.Value,
				Timestamp: item.Timestamp,
			}, true
		})
	}

	return &api.MetricQueryResult{
		Labels:     b.Labels,
		ResultType: b.ResultType,
		Values:     values,
		Value:      value,
	}
}
