package build

import (
	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/pkg/types"
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
