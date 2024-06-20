package microserverrepoimpl

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/houyi/metadata"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

func NewDatasourceMetricRepository(cli *microserver.HouYiConn) microrepository.DatasourceMetric {
	return &datasourceMetricRepositoryImpl{cli: cli}
}

type datasourceMetricRepositoryImpl struct {
	cli *microserver.HouYiConn
}

func (l *datasourceMetricRepositoryImpl) Query(ctx context.Context, req *bo.DatasourceQueryParams) ([]*bo.DatasourceQueryData, error) {
	configMap := make(map[string]string)
	if !types.TextIsNull(req.Config) {
		if err := json.Unmarshal([]byte(req.Config), &configMap); !types.IsNil(err) {
			return nil, err
		}
	}

	in := &metadata.QueryRequest{
		Query:       req.Query,
		Range:       req.TimeRange,
		Step:        req.Step,
		Endpoint:    req.Endpoint,
		Config:      configMap,
		StorageType: api.StorageType(req.StorageType),
	}
	queryReply, err := l.cli.Query(ctx, in)
	if !types.IsNil(err) {
		return nil, err
	}
	list := types.SliceTo(queryReply.GetList(), func(item *api.MetricQueryResult) *bo.DatasourceQueryData {
		var value *bo.DatasourceQueryValue
		if !types.IsNil(item.GetValue()) {
			value = &bo.DatasourceQueryValue{
				Timestamp: item.GetValue().GetTimestamp(),
				Value:     item.GetValue().GetValue(),
			}
		}
		return &bo.DatasourceQueryData{
			Labels:     item.GetLabels(),
			ResultType: item.GetResultType(),
			Values: types.SliceTo(item.GetValues(), func(item *api.MetricQueryValue) *bo.DatasourceQueryValue {
				return &bo.DatasourceQueryValue{
					Timestamp: item.GetTimestamp(),
					Value:     item.GetValue(),
				}
			}),
			Value: value,
		}
	})
	return list, nil
}

func (l *datasourceMetricRepositoryImpl) GetMetadata(ctx context.Context, datasourceInfo *bizmodel.Datasource) ([]*bizmodel.DatasourceMetric, error) {
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(datasourceInfo.Config), &configMap); !types.IsNil(err) {
		return nil, err
	}
	in := &metadata.SyncMetadataRequest{
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
