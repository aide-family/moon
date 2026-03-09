package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/merr"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

type metricDatasourceQuerier struct{}

// NewMetricDatasourceQuerier returns a MetricDatasourceQuerier that uses the metric client built from the datasource.
func NewMetricDatasourceQuerier() repository.MetricDatasourceQuerier {
	return &metricDatasourceQuerier{}
}

// ListMetrics returns metric list with name, description, unit, type only (no label queries).
func (q *metricDatasourceQuerier) ListMetrics(ctx context.Context, ds *bo.DatasourceItemBo) ([]*bo.MetricSummaryItemBo, error) {
	client, err := NewMetricClientFromDatasource(ds)
	if err != nil {
		return nil, merr.ErrorInternalServer("create metric client failed").WithCause(err)
	}
	if client == nil {
		return nil, merr.ErrorInvalidArgument("datasource is not a supported metrics type or driver")
	}

	metaResp, err := client.Metadata(ctx, "", "")
	if err != nil {
		return nil, merr.ErrorInternalServer("query metric metadata failed").WithCause(err)
	}
	if metaResp == nil {
		return nil, nil
	}

	metrics := make([]*bo.MetricSummaryItemBo, 0, len(metaResp))
	for name, item := range metaResp {
		if len(item) == 0 {
			continue
		}
		metricItem := item[0]
		metrics = append(metrics, &bo.MetricSummaryItemBo{
			Name:        name,
			Description: metricItem.Help,
			Unit:        metricItem.Unit,
			Type:        string(metricItem.Type),
		})
	}

	return metrics, nil
}

func (q *metricDatasourceQuerier) GetMetricDetail(ctx context.Context, ds *bo.DatasourceItemBo, metric string) (*bo.MetricDetailItemBo, error) {
	client, err := NewMetricClientFromDatasource(ds)
	if err != nil {
		return nil, merr.ErrorInternalServer("create metric client failed").WithCause(err)
	}

	metricDetail := &bo.MetricDetailItemBo{
		Name:   metric,
		Labels: make([]*bo.MetricLabelItemBo, 0, 10),
	}

	start := time.Time{}
	end := time.Time{}
	labelNames, _, err := client.LabelNames(ctx, []string{metric}, start, end)
	if err != nil {
		return nil, merr.ErrorInternalServer("query metric label names failed").WithCause(err)
	}

	for _, labelName := range labelNames {
		if labelName == "__name__" {
			continue
		}
		labelValues, warnings, err := client.LabelValues(ctx, labelName, []string{metric}, start, end)
		if err != nil {
			return nil, merr.ErrorInternalServer("query metric label values failed").WithCause(err)
		}
		if len(warnings) > 0 {
			klog.Warnw("msg", "query metric label values warnings", "warnings", warnings)
		}
		values := make([]string, 0, len(labelValues))
		for _, labelValue := range labelValues {
			values = append(values, string(labelValue))
		}
		metricDetail.Labels = append(metricDetail.Labels, &bo.MetricLabelItemBo{
			Name:   labelName,
			Values: values,
		})
	}

	return metricDetail, nil
}
