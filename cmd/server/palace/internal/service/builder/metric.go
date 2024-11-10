package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IMetricModuleBuilder = (*metricModuleBuilder)(nil)

type (
	metricModuleBuilder struct {
		ctx context.Context
	}

	IMetricModuleBuilder interface {
		WithUpdateMetricRequest(*datasourceapi.UpdateMetricRequest) IUpdateMetricRequestBuilder
		WithGetMetricRequest(*datasourceapi.GetMetricRequest) IGetMetricRequestBuilder
		WithListMetricRequest(*datasourceapi.ListMetricRequest) IListMetricRequestBuilder
		DoMetricBuilder() IDoMetricBuilder

		DoMetricLabelBuilder() IDoMetricLabelBuilder
	}

	IUpdateMetricRequestBuilder interface {
		ToBo() *bo.UpdateMetricParams
	}

	updateMetricRequestBuilder struct {
		ctx context.Context
		*datasourceapi.UpdateMetricRequest
	}

	IGetMetricRequestBuilder interface {
		ToBo() *bo.GetMetricParams
	}

	getMetricRequestBuilder struct {
		ctx context.Context
		*datasourceapi.GetMetricRequest
	}

	IListMetricRequestBuilder interface {
		ToBo() *bo.QueryMetricListParams
	}

	listMetricRequestBuilder struct {
		ctx context.Context
		*datasourceapi.ListMetricRequest
	}

	IDoMetricBuilder interface {
		ToAPI(*bizmodel.DatasourceMetric) *adminapi.MetricItem
		ToAPIs([]*bizmodel.DatasourceMetric) []*adminapi.MetricItem
		ToSelect(*bizmodel.DatasourceMetric) *adminapi.SelectItem
		ToSelects([]*bizmodel.DatasourceMetric) []*adminapi.SelectItem
	}

	doMetricBuilder struct {
		ctx context.Context
	}

	IDoMetricLabelBuilder interface {
		ToAPI(*bizmodel.MetricLabel) *adminapi.MetricLabelItem
		ToAPIs([]*bizmodel.MetricLabel) []*adminapi.MetricLabelItem
	}

	doMetricLabelBuilder struct {
		ctx context.Context
	}
)

func (d *doMetricLabelBuilder) ToAPI(label *bizmodel.MetricLabel) *adminapi.MetricLabelItem {
	if types.IsNil(d) || types.IsNil(label) {
		return nil
	}

	return &adminapi.MetricLabelItem{
		Name:   label.Name,
		Values: label.GetLabelValues(),
		Id:     label.ID,
	}
}

func (d *doMetricLabelBuilder) ToAPIs(labels []*bizmodel.MetricLabel) []*adminapi.MetricLabelItem {
	if types.IsNil(d) || types.IsNil(labels) {
		return nil
	}

	return types.SliceTo(labels, func(label *bizmodel.MetricLabel) *adminapi.MetricLabelItem {
		return d.ToAPI(label)
	})
}

func (m *metricModuleBuilder) DoMetricLabelBuilder() IDoMetricLabelBuilder {
	return &doMetricLabelBuilder{ctx: m.ctx}
}

func (d *doMetricBuilder) ToAPI(metric *bizmodel.DatasourceMetric) *adminapi.MetricItem {
	if types.IsNil(d) || types.IsNil(metric) {
		return nil
	}

	return &adminapi.MetricItem{
		Name:       metric.Name,
		Help:       metric.Remark,
		Type:       api.MetricType(metric.Category),
		Labels:     NewParamsBuild(d.ctx).MetricModuleBuilder().DoMetricLabelBuilder().ToAPIs(metric.Labels),
		Unit:       metric.Unit,
		Id:         metric.ID,
		LabelCount: metric.LabelCount,
	}
}

func (d *doMetricBuilder) ToAPIs(metrics []*bizmodel.DatasourceMetric) []*adminapi.MetricItem {
	if types.IsNil(d) || types.IsNil(metrics) {
		return nil
	}

	return types.SliceTo(metrics, func(metric *bizmodel.DatasourceMetric) *adminapi.MetricItem {
		return d.ToAPI(metric)
	})
}

func (d *doMetricBuilder) ToSelect(metric *bizmodel.DatasourceMetric) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(metric) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    metric.ID,
		Label:    metric.Name,
		Disabled: metric.DeletedAt > 0,
		Extend: &adminapi.SelectExtend{
			Remark: metric.Remark,
		},
	}
}

func (d *doMetricBuilder) ToSelects(metrics []*bizmodel.DatasourceMetric) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(metrics) {
		return nil
	}

	return types.SliceTo(metrics, func(metric *bizmodel.DatasourceMetric) *adminapi.SelectItem {
		return d.ToSelect(metric)
	})
}

func (l *listMetricRequestBuilder) ToBo() *bo.QueryMetricListParams {
	if types.IsNil(l) {
		return nil
	}

	return &bo.QueryMetricListParams{
		Page:         types.NewPagination(l.GetPagination()),
		Keyword:      l.GetKeyword(),
		DatasourceID: l.GetDatasourceId(),
		MetricType:   vobj.MetricType(l.GetMetricType()),
	}
}

func (g *getMetricRequestBuilder) ToBo() *bo.GetMetricParams {
	if types.IsNil(g) || types.IsNil(g.GetMetricRequest) {
		return nil
	}

	return &bo.GetMetricParams{
		ID:           g.GetId(),
		WithRelation: true,
	}
}

func (u *updateMetricRequestBuilder) ToBo() *bo.UpdateMetricParams {
	if types.IsNil(u) || types.IsNil(u.UpdateMetricRequest) {
		return nil
	}

	return &bo.UpdateMetricParams{
		ID:     u.GetId(),
		Remark: u.GetRemark(),
		Unit:   u.GetUnit(),
	}
}

func (m *metricModuleBuilder) WithUpdateMetricRequest(request *datasourceapi.UpdateMetricRequest) IUpdateMetricRequestBuilder {
	return &updateMetricRequestBuilder{ctx: m.ctx, UpdateMetricRequest: request}
}

func (m *metricModuleBuilder) WithGetMetricRequest(request *datasourceapi.GetMetricRequest) IGetMetricRequestBuilder {
	return &getMetricRequestBuilder{ctx: m.ctx, GetMetricRequest: request}
}

func (m *metricModuleBuilder) WithListMetricRequest(request *datasourceapi.ListMetricRequest) IListMetricRequestBuilder {
	return &listMetricRequestBuilder{ctx: m.ctx, ListMetricRequest: request}
}

func (m *metricModuleBuilder) DoMetricBuilder() IDoMetricBuilder {
	return &doMetricBuilder{ctx: m.ctx}
}
