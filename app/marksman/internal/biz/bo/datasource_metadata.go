package bo

import (
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// MetricSummaryItemBo is a metric list item: name, help, unit, type (no labels). Help aligns with Prometheus metadata.
type MetricSummaryItemBo struct {
	Name   string
	Help   string
	Unit   string
	Type   string
}

// ToAPIV1ListMetricsReply converts BO to proto.
func ToAPIV1ListMetricsReply(rs []*MetricSummaryItemBo) *apiv1.ListMetricsReply {
	if rs == nil {
		return &apiv1.ListMetricsReply{}
	}
	out := make([]*apiv1.MetricSummaryItem, 0, len(rs))
	for _, m := range rs {
		out = append(out, &apiv1.MetricSummaryItem{
			Name: m.Name,
			Help: m.Help,
			Unit: m.Unit,
			Type: m.Type,
		})
	}
	return &apiv1.ListMetricsReply{Metrics: out}
}

// MetricDetailItemBo is one metric's full metadata including labels and label values.
type MetricDetailItemBo struct {
	Name   string
	Help   string
	Unit   string
	Type   string
	Labels []*MetricLabelItemBo
}

type MetricLabelItemBo struct {
	Name   string
	Values []string
}

func ToAPIV1MetricLabelItem(labels []*MetricLabelItemBo) []*apiv1.MetricLabelItem {
	if labels == nil {
		return nil
	}
	out := make([]*apiv1.MetricLabelItem, 0, len(labels))
	for _, l := range labels {
		out = append(out, &apiv1.MetricLabelItem{
			Name:   l.Name,
			Values: l.Values,
		})
	}
	return out
}

// ToAPIV1GetMetricDetailReply converts BO to proto.
func ToAPIV1GetMetricDetailReply(r *MetricDetailItemBo) *apiv1.MetricDetailItem {
	if r == nil {
		return &apiv1.MetricDetailItem{}
	}
	return &apiv1.MetricDetailItem{
		Name:   r.Name,
		Help:   r.Help,
		Unit:   r.Unit,
		Type:   r.Type,
		Labels: ToAPIV1MetricLabelItem(r.Labels),
	}
}
