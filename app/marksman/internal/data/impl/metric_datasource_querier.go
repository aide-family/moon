package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/merr"
	klog "github.com/go-kratos/kratos/v2/log"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

// NewMetricDatasourceQuerierRepository returns a MetricDatasourceQuerier that uses the metric client built from the datasource.
func NewMetricDatasourceQuerierRepository() repository.MetricDatasourceQuerier {
	return &metricDatasourceQuerierRepository{}
}

type metricDatasourceQuerierRepository struct{}

// ListMetrics returns metric list with name, help, unit, type only (no label queries).
func (q *metricDatasourceQuerierRepository) ListMetrics(ctx context.Context, ds *bo.DatasourceItemBo) ([]*bo.MetricSummaryItemBo, error) {
	client, err := NewMetricClientFromDatasource(ds)
	if err != nil {
		return nil, merr.ErrorInternalServer("create metric client failed").WithCause(err)
	}
	if client == nil {
		return nil, merr.ErrorInvalidArgument("datasource is not a supported metrics type or driver")
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Hour * 24 * 30)
	// Use LabelValues as source of truth (same as Prometheus UI /api/v1/label/__name__/values).
	// Pass zero start/end so the client does not add start/end params; then Prometheus returns
	// all label values (like the UI) instead of filtering by time range.
	allMetricNames, _, err := client.LabelValues(ctx, "__name__", []string{}, startTime, endTime)
	if err != nil {
		return nil, merr.ErrorInternalServer("query metric label names failed").WithCause(err)
	}
	klog.Debugw("msg", "list metrics label values", "count", len(allMetricNames), "allMetricNames", allMetricNames)

	// Enrich with metadata when available; do not fail or return empty when metadata is nil.
	metaResp, _ := client.Metadata(ctx, "", "")

	metrics := make([]*bo.MetricSummaryItemBo, 0, len(allMetricNames))
	for _, name := range allMetricNames {
		item := &bo.MetricSummaryItemBo{
			Name: string(name),
		}
		if metaResp != nil {
			metricMetadata, ok := metaResp[string(name)]
			if ok && len(metricMetadata) > 0 {
				item.Help = metricMetadata[0].Help
				item.Unit = metricMetadata[0].Unit
				item.Type = string(metricMetadata[0].Type)
			}
		}
		metrics = append(metrics, item)
	}

	return metrics, nil
}

func (q *metricDatasourceQuerierRepository) GetMetricDetail(ctx context.Context, ds *bo.DatasourceItemBo, metric string) (*bo.MetricDetailItemBo, error) {
	client, err := NewMetricClientFromDatasource(ds)
	if err != nil {
		return nil, merr.ErrorInternalServer("create metric client failed").WithCause(err)
	}

	metricMetadatas, err := client.Metadata(ctx, metric, "1")
	if err != nil {
		return nil, merr.ErrorInternalServer("query metric metadata failed").WithCause(err)
	}
	klog.Debugw("msg", "query metric metadata", "metric", metric, "metricMetadatas", metricMetadatas)

	metricDetail := &bo.MetricDetailItemBo{
		Name:   metric,
		Labels: make([]*bo.MetricLabelItemBo, 0, 10),
	}

	if metricMetadata, ok := metricMetadatas[metric]; ok && len(metricMetadata) > 0 {
		metricDetail.Help = metricMetadata[0].Help
		metricDetail.Unit = metricMetadata[0].Unit
		metricDetail.Type = string(metricMetadata[0].Type)
	}

	end := time.Now()
	start := end.Add(-time.Hour * 24 * 30)
	labelNames, _, err := client.LabelNames(ctx, []string{metric}, start, end)
	if err != nil {
		return nil, merr.ErrorInternalServer("query metric label names failed").WithCause(err)
	}
	klog.Debugw("msg", "query metric label names", "metric", metric, "labelNames", labelNames)

	for _, labelName := range labelNames {
		if labelName == "__name__" {
			continue
		}
		labelValues, _, err := client.LabelValues(ctx, labelName, []string{}, start, end)
		if err != nil {
			return nil, merr.ErrorInternalServer("query metric label values failed").WithCause(err)
		}
		klog.Debugw("msg", "query metric label values", "metric", metric, "labelName", labelName, "labelValues", labelValues)
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

// QueryRange implements [repository.MetricDatasourceQuerier].
func (q *metricDatasourceQuerierRepository) QueryRange(ctx context.Context, ds *bo.DatasourceItemBo, query string, queryRange v1.Range) (model.Matrix, error) {
	client, err := NewMetricClientFromDatasource(ds)
	if err != nil {
		return nil, err
	}

	value, _, err := client.QueryRange(ctx, query, queryRange)
	if err != nil {
		return nil, err
	}
	matrix, ok := value.(model.Matrix)
	if !ok || len(matrix) == 0 {
		return model.Matrix{}, nil
	}
	return matrix, nil
}
