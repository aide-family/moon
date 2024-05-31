package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/data"
	"github.com/aide-cloud/moon/pkg/datasource/metric"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
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
