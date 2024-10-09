package datasource

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
	"github.com/go-kratos/kratos/v2/log"
)

type (
	// Value 数据源查询值
	Value struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}

	// Point 数据点
	Point struct {
		// 标签集合
		Labels map[string]string `json:"labels"`
		// 值
		Values []*Value `json:"value"`
	}
)

var _ Datasource = (*datasource)(nil)

// Datasource 数据源通用接口
type Datasource interface {
	Metric() (MetricDatasource, error)
}

type datasource struct {
	config *api.Datasource
}

func (d *datasource) Metric() (MetricDatasource, error) {
	if types.IsNil(d) || types.IsNil(d.config) {
		return nil, merr.ErrorNotificationSystemError("datasource is nil")
	}

	dataType := vobj.DatasourceType(d.config.GetCategory())
	if !dataType.IsMetrics() {
		return nil, merr.ErrorNotificationSystemError("not a metric datasource: %s", dataType)
	}
	opts := []MetricDatasourceBuildOption{
		WithMetricID(d.config.GetId()),
		WithMetricStep(10),
		WithMetricEndpoint(d.config.GetEndpoint()),
		WithMetricBasicAuth(d.config.GetConfig()["username"], d.config.GetConfig()["password"]),
	}
	return NewMetricDatasource(vobj.StorageType(d.config.GetStorageType()), opts...)
}

// NewDatasource 根据配置创建对应的数据源
func NewDatasource(config *api.Datasource) Datasource {
	return &datasource{config: config}
}

type EvalFunc func(ctx context.Context, expr string, duration *types.Duration) (map[watch.Indexer]*Point, error)

func MetricEval(items ...MetricDatasource) EvalFunc {
	return func(ctx context.Context, expr string, duration *types.Duration) (map[watch.Indexer]*Point, error) {
		evalRes := make(map[watch.Indexer]*Point)
		endAt := time.Now()
		startAt := types.NewTime(endAt.Add(-duration.Duration.AsDuration()))
		for _, item := range items {
			list, err := metricEval(ctx, item, expr, startAt.Unix(), endAt.Unix())
			if err != nil {
				log.Warnw("eval", err)
				continue
			}
			for k, v := range list {
				evalRes[k] = v
			}
		}
		return evalRes, nil
	}
}

func metricEval(ctx context.Context, d MetricDatasource, expr string, startAt, endAt int64) (map[watch.Indexer]*Point, error) {
	queryRange, err := d.QueryRange(ctx, expr, startAt, endAt, d.Step())
	if err != nil {
		return nil, err
	}
	basicInfo := d.GetBasicInfo()
	var responseMap = make(map[watch.Indexer]*Point)
	for _, response := range queryRange {
		labels := response.Labels
		values := make([]*Value, 0, len(response.Values))
		for _, v := range response.Values {
			values = append(values, &Value{
				Value:     v.Value,
				Timestamp: v.Timestamp,
			})
		}
		labels[vobj.DatasourceID] = strconv.Itoa(int(basicInfo.ID))
		labels[vobj.DatasourceURL] = basicInfo.Endpoint
		vobjLabels := vobj.NewLabels(labels)
		responseMap[vobjLabels] = &Point{
			Values: values,
			Labels: labels,
		}
	}
	return responseMap, nil
}
