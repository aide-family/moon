package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data/microserver"
	"github.com/aide-family/moon/pkg/houyi/datasource/metric"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewMetricRepository 实例化MetricRepository
func NewMetricRepository(data *data.Data, palaceCli *microserver.PalaceConn) repository.Metric {
	return &metricRepositoryImpl{data: data, palaceCli: palaceCli}
}

type metricRepositoryImpl struct {
	data      *data.Data
	palaceCli *microserver.PalaceConn
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

// GetMetrics 获取指标列表
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

	metadataMap := make(map[string]*metric.Metric)
	for _, item := range metadata.Metric {
		metadataMap[item.Name] = item
	}
	list := make([]*bo.MetricDetail, 0, len(metadata.Metric))
	for _, metricDetail := range metadataMap {
		list = append(list, &bo.MetricDetail{
			Name:   metricDetail.Name,
			Help:   metricDetail.Help,
			Type:   metricDetail.Type,
			Unit:   metricDetail.Unit,
			Labels: metricDetail.Labels,
		})
	}
	return list, nil
}

// Query 查询指标
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

// PushMetric 推送指标
func (l *metricRepositoryImpl) PushMetric(ctx context.Context, req *bo.PushMetricParams) error {
	labels := make([]*admin.MetricLabel, 0, len(req.Labels))
	for label, labelValue := range req.Labels {
		labels = append(labels, &admin.MetricLabel{
			Name: label,
			Values: types.SliceTo(labelValue, func(item string) *admin.MetricLabelValue {
				return &admin.MetricLabelValue{
					Value: item,
				}
			}),
		})
	}

	_, err := l.palaceCli.PushMetric(ctx, &datasourceapi.SyncMetricRequest{
		Metrics: &admin.MetricDetail{
			Name:   req.Name,
			Help:   req.Help,
			Type:   api.MetricType(vobj.GetMetricType(req.Type)),
			Labels: labels,
			Unit:   req.Unit,
		},
		Done:         req.Done,
		DatasourceId: req.DatasourceID,
		TeamId:       req.TeamID,
	})
	return err
}
