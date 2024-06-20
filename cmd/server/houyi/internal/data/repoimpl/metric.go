package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func NewMetricRepository(data *data.Data) repository.Metric {
	return &metricRepositoryImpl{data: data}
}

type metricRepositoryImpl struct {
	data *data.Data
}

func (l *metricRepositoryImpl) getMetricOptions(datasourceInfo *bo.GetMetricsParams) ([]metric.DatasourceBuildOption, error) {
	var opts []metric.DatasourceBuildOption
	switch datasourceInfo.StorageType {
	case vobj.StorageTypePrometheus:
		opts = append(opts, metric.WithPrometheusOption(
			metric.WithPrometheusEndpoint(datasourceInfo.Endpoint),
			metric.WithPrometheusConfig(datasourceInfo.Config),
		))
	default:
		return nil, merr.ErrorNotification("不支持的存储类型").WithMetadata(map[string]string{
			"storage_type": datasourceInfo.StorageType.String(),
		})
	}
	return opts, nil
}

func (l *metricRepositoryImpl) GetMetrics(ctx context.Context, datasourceInfo *bo.GetMetricsParams) ([]*bo.MetricDetail, error) {
	opts, err := l.getMetricOptions(datasourceInfo)
	if err != nil {
		return nil, err
	}
	datasource, err := metric.NewMetricDatasource(datasourceInfo.StorageType, opts...)
	if err != nil {
		return nil, err
	}
	metadata, err := datasource.Metadata(ctx)
	if err != nil {
		return nil, err
	}

	list := types.SliceTo(metadata.Metric, func(item *metric.Metric) *bo.MetricDetail {
		return &bo.MetricDetail{Name: item.Name, Help: item.Help, Type: item.Type, Labels: item.Labels, Unit: item.Unit}
	})
	return list, nil
}

func (l *metricRepositoryImpl) Query(ctx context.Context, req *bo.QueryQLParams) ([]*metric.QueryResponse, error) {
	opts, err := l.getMetricOptions(&req.GetMetricsParams)
	if err != nil {
		return nil, err
	}
	datasource, err := metric.NewMetricDatasource(req.StorageType, opts...)
	if err != nil {
		return nil, err
	}
	if len(req.TimeRange) == 0 {
		return nil, merr.ErrorNotification("time range is empty")
	}
	if len(req.TimeRange) == 1 {
		return datasource.Query(ctx, req.QueryQL, types.NewTimeByString(req.TimeRange[0]).Unix())
	}
	start, end := types.NewTimeByString(req.TimeRange[0]).Unix(), types.NewTimeByString(req.TimeRange[1]).Unix()
	return datasource.QueryRange(ctx, req.QueryQL, start, end, req.Step)
}
