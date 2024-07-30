package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewMetricBuilder 创建Metric构造器
func NewMetricBuilder(metricDetail *bo.MetricDetail) *MetricBuilder {
	return &MetricBuilder{
		MetricDetail: metricDetail,
	}
}

// MetricBuilder 构建Metric对象
type MetricBuilder struct {
	*bo.MetricDetail
}

// ToAPI 转换为api对象
func (b *MetricBuilder) ToAPI() *api.MetricDetail {
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
		Type:   api.MetricType(vobj.GetMetricType(b.Type)),
		Labels: labels,
		Unit:   b.Unit,
	}
}

// NewMetricQueryBuilder 创建MetricQuery构造器
func NewMetricQueryBuilder(queryResponse *metric.QueryResponse) *MetricQueryBuilder {
	return &MetricQueryBuilder{
		QueryResponse: queryResponse,
	}
}

// MetricQueryBuilder 构建MetricQuery对象
type MetricQueryBuilder struct {
	*metric.QueryResponse
}

// ToAPI 转换为api对象
func (b *MetricQueryBuilder) ToAPI() *api.MetricQueryResult {
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
		Labels:     b.Labels.Map(),
		ResultType: b.ResultType,
		Values:     values,
		Value:      value,
	}
}
