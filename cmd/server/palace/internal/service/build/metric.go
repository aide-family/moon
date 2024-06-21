package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type DatasourceMetricBuilder struct {
	*bizmodel.DatasourceMetric
}

func NewDatasourceMetricBuilder(metric *bizmodel.DatasourceMetric) *DatasourceMetricBuilder {
	return &DatasourceMetricBuilder{DatasourceMetric: metric}
}

// ToApi 转换为api对象
func (b *DatasourceMetricBuilder) ToApi() *admin.MetricDetail {
	if types.IsNil(b) || types.IsNil(b.DatasourceMetric) {
		return nil
	}

	return &admin.MetricDetail{
		Name: b.Name,
		Help: b.Remark,
		Type: api.MetricType(b.Category),
		Labels: types.SliceTo(b.Labels, func(item *bizmodel.MetricLabel) *admin.MetricLabel {
			return NewDatasourceMetricLabelBuilder(item).ToApi()
		}),
		Unit: b.Unit,
		Id:   b.ID,
	}
}

type DatasourceMetricLabelBuilder struct {
	*bizmodel.MetricLabel
}

func NewDatasourceMetricLabelBuilder(label *bizmodel.MetricLabel) *DatasourceMetricLabelBuilder {
	return &DatasourceMetricLabelBuilder{MetricLabel: label}
}

// ToApi 转换为api对象
func (b *DatasourceMetricLabelBuilder) ToApi() *admin.MetricLabel {
	if types.IsNil(b) || types.IsNil(b.MetricLabel) {
		return nil
	}

	return &admin.MetricLabel{
		Name: b.Name,
		Values: types.SliceTo(b.LabelValues, func(item *bizmodel.MetricLabelValue) *admin.MetricLabelValue {
			return NewDatasourceMetricLabelValueBuilder(item).ToApi()
		}),
		Id: b.ID,
	}
}

type DatasourceMetricLabelValueBuilder struct {
	*bizmodel.MetricLabelValue
}

func NewDatasourceMetricLabelValueBuilder(value *bizmodel.MetricLabelValue) *DatasourceMetricLabelValueBuilder {
	return &DatasourceMetricLabelValueBuilder{MetricLabelValue: value}
}

// ToApi 转换为api对象
func (b *DatasourceMetricLabelValueBuilder) ToApi() *admin.MetricLabelValue {
	if types.IsNil(b) || types.IsNil(b.MetricLabelValue) {
		return nil
	}

	return &admin.MetricLabelValue{
		Id:    b.ID,
		Value: b.Name,
	}
}
