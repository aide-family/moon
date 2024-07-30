package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type (
	// DatasourceMetricModelBuilder 数据源指标模型转换器
	DatasourceMetricModelBuilder interface {
		ToAPI() *admin.MetricDetail
	}

	datasourceMetricModelBuilder struct {
		*bizmodel.DatasourceMetric

		ctx context.Context
	}

	// DatasourceMetricLabelModelBuilder 数据源指标标签模型转换器
	DatasourceMetricLabelModelBuilder interface {
		ToAPI() *admin.MetricLabel
	}

	datasourceMetricLabelModelBuilder struct {
		*bizmodel.MetricLabel
		ctx context.Context
	}

	// DatasourceMetricLabelValueBuilder 数据源指标标签值模型转换器
	DatasourceMetricLabelValueBuilder interface {
		ToAPI() *admin.MetricLabelValue
	}

	datasourceMetricLabelValueBuilder struct {
		*bizmodel.MetricLabelValue
		ctx context.Context
	}
)

// ToAPI 转换为api对象
func (b *datasourceMetricModelBuilder) ToAPI() *admin.MetricDetail {
	if types.IsNil(b) || types.IsNil(b.DatasourceMetric) {
		return nil
	}

	return &admin.MetricDetail{
		Name: b.Name,
		Help: b.Remark,
		Type: api.MetricType(b.Category),
		Labels: types.SliceTo(b.Labels, func(item *bizmodel.MetricLabel) *admin.MetricLabel {
			return NewBuilder().WithAPIDatasourceMetricLabel(item).ToAPI()
		}),
		Unit: b.Unit,
		Id:   b.ID,
	}
}

// ToAPI 转换为api对象
func (b *datasourceMetricLabelModelBuilder) ToAPI() *admin.MetricLabel {
	if types.IsNil(b) || types.IsNil(b.MetricLabel) {
		return nil
	}

	return &admin.MetricLabel{
		Name: b.Name,
		Values: types.SliceTo(b.LabelValues, func(item *bizmodel.MetricLabelValue) *admin.MetricLabelValue {
			return NewBuilder().WithAPIDatasourceMetricLabelValue(item).ToAPI()
		}),
		Id: b.ID,
	}
}

// ToAPI 转换为api对象
func (b *datasourceMetricLabelValueBuilder) ToAPI() *admin.MetricLabelValue {
	if types.IsNil(b) || types.IsNil(b.MetricLabelValue) {
		return nil
	}

	return &admin.MetricLabelValue{
		Id:    b.ID,
		Value: b.Name,
	}
}
