package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type (
	DatasourceMetricModelBuilder interface {
		ToApi() *admin.MetricDetail
	}
	datasourceMetricModelBuilder struct {
		*bizmodel.DatasourceMetric

		ctx context.Context
	}

	DatasourceMetricLabelModelBuilder interface {
		ToApi() *admin.MetricLabel
	}
	datasourceMetricLabelModelBuilder struct {
		*bizmodel.MetricLabel
		ctx context.Context
	}

	DatasourceMetricLabelValueBuilder interface {
		ToApi() *admin.MetricLabelValue
	}
	datasourceMetricLabelValueBuilder struct {
		*bizmodel.MetricLabelValue
		ctx context.Context
	}
)

// ToApi 转换为api对象
func (b *datasourceMetricModelBuilder) ToApi() *admin.MetricDetail {
	if types.IsNil(b) || types.IsNil(b.DatasourceMetric) {
		return nil
	}

	return &admin.MetricDetail{
		Name: b.Name,
		Help: b.Remark,
		Type: api.MetricType(b.Category),
		Labels: types.SliceTo(b.Labels, func(item *bizmodel.MetricLabel) *admin.MetricLabel {
			return NewBuilder().WithApiDatasourceMetricLabel(item).ToApi()
		}),
		Unit: b.Unit,
		Id:   b.ID,
	}
}

// ToApi 转换为api对象
func (b *datasourceMetricLabelModelBuilder) ToApi() *admin.MetricLabel {
	if types.IsNil(b) || types.IsNil(b.MetricLabel) {
		return nil
	}

	return &admin.MetricLabel{
		Name: b.Name,
		Values: types.SliceTo(b.LabelValues, func(item *bizmodel.MetricLabelValue) *admin.MetricLabelValue {
			return NewBuilder().WithApiDatasourceMetricLabelValue(item).ToApi()
		}),
		Id: b.ID,
	}
}

// ToApi 转换为api对象
func (b *datasourceMetricLabelValueBuilder) ToApi() *admin.MetricLabelValue {
	if types.IsNil(b) || types.IsNil(b.MetricLabelValue) {
		return nil
	}

	return &admin.MetricLabelValue{
		Id:    b.ID,
		Value: b.Name,
	}
}
