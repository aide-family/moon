package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
)

type DatasourceMetricBuild struct {
	*bizmodel.DatasourceMetric
}

func NewDatasourceMetricBuild(metric *bizmodel.DatasourceMetric) *DatasourceMetricBuild {
	return &DatasourceMetricBuild{DatasourceMetric: metric}
}

// ToApi 转换为api对象
func (b *DatasourceMetricBuild) ToApi() *admin.MetricDetail {
	if types.IsNil(b) || types.IsNil(b.DatasourceMetric) {
		return nil
	}

	return &admin.MetricDetail{
		Name: b.Name,
		Help: b.Remark,
		Type: api.MetricType(b.Category),
		Labels: types.SliceTo(b.Labels, func(item *bizmodel.MetricLabel) *admin.MetricLabel {
			return NewDatasourceMetricLabelBuild(item).ToApi()
		}),
		Unit: b.Unit,
		Id:   b.ID,
	}
}

type DatasourceMetricLabelBuild struct {
	*bizmodel.MetricLabel
}

func NewDatasourceMetricLabelBuild(label *bizmodel.MetricLabel) *DatasourceMetricLabelBuild {
	return &DatasourceMetricLabelBuild{MetricLabel: label}
}

// ToApi 转换为api对象
func (b *DatasourceMetricLabelBuild) ToApi() *admin.MetricLabel {
	if types.IsNil(b) || types.IsNil(b.MetricLabel) {
		return nil
	}

	return &admin.MetricLabel{
		Name: b.Name,
		Values: types.SliceTo(b.LabelValues, func(item *bizmodel.MetricLabelValue) *admin.MetricLabelValue {
			return NewDatasourceMetricLabelValueBuild(item).ToApi()
		}),
		Id: b.ID,
	}
}

type DatasourceMetricLabelValueBuild struct {
	*bizmodel.MetricLabelValue
}

func NewDatasourceMetricLabelValueBuild(value *bizmodel.MetricLabelValue) *DatasourceMetricLabelValueBuild {
	return &DatasourceMetricLabelValueBuild{MetricLabelValue: value}
}

// ToApi 转换为api对象
func (b *DatasourceMetricLabelValueBuild) ToApi() *admin.MetricLabelValue {
	if types.IsNil(b) || types.IsNil(b.MetricLabelValue) {
		return nil
	}

	return &admin.MetricLabelValue{
		Id:    b.ID,
		Value: b.Name,
	}
}
