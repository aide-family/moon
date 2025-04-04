package datasource

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/label"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/log"
)

type (
	// Value 数据源查询值
	Value struct {
		Value     float64        `json:"value"`
		Timestamp int64          `json:"timestamp"`
		Ext       map[string]any `json:"ext"`
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
	config *api.DatasourceItem
}

func (d *datasource) Metric() (MetricDatasource, error) {
	if types.IsNil(d) || types.IsNil(d.config) {
		return nil, merr.ErrorNotificationSystemError("datasource is nil")
	}

	dataType := vobj.DatasourceType(d.config.GetCategory())
	if !dataType.IsMetrics() {
		return nil, merr.ErrorNotificationSystemError("not a metric datasource: %s", dataType)
	}
	config := make(map[string]any)
	_ = types.Unmarshal([]byte(d.config.GetConfig()), &config)
	username, _ := config["username"].(string)
	password, _ := config["password"].(string)
	opts := []MetricDatasourceBuildOption{
		WithMetricID(d.config.GetId()),
		WithMetricStep(10),
		WithMetricEndpoint(d.config.GetEndpoint()),
		WithMetricBasicAuth(username, password),
	}
	return NewMetricDatasource(vobj.StorageType(d.config.GetStorageType()), opts...)
}

// NewDatasource 根据配置创建对应的数据源
func NewDatasource(config *api.DatasourceItem) Datasource {
	return &datasource{config: config}
}

// EvalFunc 指标评估函数
type EvalFunc func(ctx context.Context, expr string, duration *types.Duration) (map[watch.Indexer]*Point, error)

// MetricEval 指标评估函数
func MetricEval(items ...MetricDatasource) EvalFunc {
	return func(ctx context.Context, expr string, duration *types.Duration) (map[watch.Indexer]*Point, error) {
		var wg sync.WaitGroup
		evalRes := safety.NewMap[watch.Indexer, *Point]()
		endAt := time.Now()
		startAt := types.NewTime(endAt.Add(-duration.Duration.AsDuration()))
		for _, item := range items {
			wg.Add(1)
			go func(d MetricDatasource) {
				defer after.RecoverX()
				defer wg.Done()
				list, err := metricEval(ctx, item, expr, startAt.Unix(), endAt.Unix())
				if err != nil {
					log.Warnw("eval", err)
					return
				}
				for k, v := range list {
					evalRes.Set(k, v)
				}
			}(item)
		}
		wg.Wait()
		res := make(map[watch.Indexer]*Point)
		for key, value := range evalRes.List() {
			res[key] = value
		}
		return res, nil
	}
}

// metricEval 指标评估
func metricEval(ctx context.Context, d MetricDatasource, expr string, startAt, endAt int64) (map[watch.Indexer]*Point, error) {
	queryRange, err := d.QueryRange(ctx, expr, startAt, endAt, d.Step())
	if err != nil {
		return nil, err
	}
	basicInfo := d.GetBasicInfo()
	responseMap := make(map[watch.Indexer]*Point)
	for _, response := range queryRange {
		labels := response.Labels
		values := make([]*Value, 0, len(response.Values))
		for _, v := range response.Values {
			values = append(values, &Value{
				Value:     v.Value,
				Timestamp: v.Timestamp,
			})
		}
		labels[label.DatasourceID] = strconv.Itoa(int(basicInfo.ID))
		labels[label.DatasourceURL] = basicInfo.Endpoint
		vobjLabels := label.NewLabels(labels)
		responseMap[vobjLabels] = &Point{
			Values: values,
			Labels: labels,
		}
	}
	return responseMap, nil
}
