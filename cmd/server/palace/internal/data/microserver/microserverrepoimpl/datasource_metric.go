package microserverrepoimpl

import (
	"context"
	"encoding/json"

	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/houyi/metadata"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewDatasourceMetricRepository(cli *microserver.HouYiConn) microrepository.DatasourceMetric {
	return &datasourceMetricRepositoryImpl{cli: cli}
}

type datasourceMetricRepositoryImpl struct {
	cli *microserver.HouYiConn
}

func (l *datasourceMetricRepositoryImpl) GetMetadata(ctx context.Context, datasourceInfo *bizmodel.Datasource) ([]*bizmodel.DatasourceMetric, error) {
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(datasourceInfo.Config), &configMap); !types.IsNil(err) {
		return nil, err
	}
	in := &metadata.SyncRequest{
		Endpoint:    datasourceInfo.Endpoint,
		Config:      configMap,
		StorageType: api.StorageType(datasourceInfo.StorageType),
	}
	syncReply, err := l.cli.Sync(ctx, in)
	if !types.IsNil(err) {
		return nil, err
	}
	metrics := make([]*bizmodel.DatasourceMetric, 0, len(syncReply.GetMetrics()))
	for _, metric := range syncReply.GetMetrics() {
		labels := make([]*bizmodel.MetricLabel, 0, len(metric.GetLabels()))
		for labelName, labelValues := range metric.GetLabels() {
			labels = append(labels, &bizmodel.MetricLabel{
				Name: labelName,
				LabelValues: types.SliceTo(labelValues.GetValues(), func(val string) *bizmodel.MetricLabelValue {
					return &bizmodel.MetricLabelValue{
						Name: val,
					}
				}),
			})
		}
		item := &bizmodel.DatasourceMetric{
			Name:         metric.GetName(),
			Category:     getMetricType(metric.GetType()),
			Unit:         metric.GetUnit(),
			Remark:       metric.GetHelp(),
			DatasourceID: datasourceInfo.ID,
			Labels:       labels,
		}
		metrics = append(metrics, item)
	}

	return metrics, nil
}

// getMetricType 获取指标类型
func getMetricType(metricType string) vobj.MetricType {
	switch metricType {
	case "counter":
		return vobj.MetricTypeCounter
	case "histogram":
		return vobj.MetricTypeHistogram
	case "gauge":
		return vobj.MetricTypeGauge
	case "summary":
		return vobj.MetricTypeSummary
	default:
		return vobj.MetricTypeUnknown
	}
}
