package impl

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/laurel/internal/data"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/slices"
)

func NewCacheRepo(d *data.Data, logger log.Logger) repository.Cache {
	return &cacheImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.cache")),
	}
}

type cacheImpl struct {
	*data.Data

	helper *log.Helper
}

// StorageMetric implements repository.Cache.
func (c *cacheImpl) StorageMetric(ctx context.Context, metrics ...bo.MetricVec) error {
	metricsByType := slices.GroupBy(metrics, func(metric bo.MetricVec) vobj.MetricType {
		return metric.Type()
	})
	counterMetrics := metricsByType[vobj.MetricTypeCounter]
	gaugeMetrics := metricsByType[vobj.MetricTypeGauge]
	histogramMetrics := metricsByType[vobj.MetricTypeHistogram]
	summaryMetrics := metricsByType[vobj.MetricTypeSummary]

	if len(counterMetrics) > 0 {
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeCounter)
		values := make(map[string]any, len(counterMetrics))
		for _, metric := range counterMetrics {
			values[metric.GetMetricName()] = metric
		}
		if err := c.GetCache().Client().HSet(ctx, key, values).Err(); err != nil {
			c.helper.WithContext(ctx).Errorw("method", "counterMetrics.HSet", "err", err)
		}
	}
	if len(gaugeMetrics) > 0 {
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeGauge)
		values := make(map[string]any, len(gaugeMetrics))
		for _, metric := range gaugeMetrics {
			values[metric.GetMetricName()] = metric
		}
		if err := c.GetCache().Client().HSet(ctx, key, values).Err(); err != nil {
			c.helper.WithContext(ctx).Errorw("method", "gaugeMetrics.HSet", "err", err)
		}
	}
	if len(histogramMetrics) > 0 {
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeHistogram)
		values := make(map[string]any, len(histogramMetrics))
		for _, metric := range histogramMetrics {
			values[metric.GetMetricName()] = metric
		}
		if err := c.GetCache().Client().HSet(ctx, key, values).Err(); err != nil {
			c.helper.WithContext(ctx).Errorw("method", "histogramMetrics.HSet", "err", err)
		}
	}
	if len(summaryMetrics) > 0 {
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeSummary)
		values := make(map[string]any, len(summaryMetrics))
		for _, metric := range summaryMetrics {
			values[metric.GetMetricName()] = metric
		}
		if err := c.GetCache().Client().HSet(ctx, key, values).Err(); err != nil {
			c.helper.WithContext(ctx).Errorw("method", "summaryMetrics.HSet", "err", err)
		}
	}

	return nil
}

func (c *cacheImpl) GetCounterMetrics(ctx context.Context, names ...string) ([]*bo.CounterMetricVec, error) {
	key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeCounter)
	return getMetrics[bo.CounterMetricVec](ctx, c.GetCache(), key, names...)
}

func (c *cacheImpl) GetGaugeMetrics(ctx context.Context, names ...string) ([]*bo.GaugeMetricVec, error) {
	key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeGauge)
	return getMetrics[bo.GaugeMetricVec](ctx, c.GetCache(), key, names...)
}

func (c *cacheImpl) GetHistogramMetrics(ctx context.Context, names ...string) ([]*bo.HistogramMetricVec, error) {
	key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeHistogram)
	return getMetrics[bo.HistogramMetricVec](ctx, c.GetCache(), key, names...)
}

func (c *cacheImpl) GetSummaryMetrics(ctx context.Context, names ...string) ([]*bo.SummaryMetricVec, error) {
	key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeSummary)
	return getMetrics[bo.SummaryMetricVec](ctx, c.GetCache(), key, names...)
}

func getMetrics[T any](ctx context.Context, cacheInstance cache.Cache, key string, names ...string) ([]*T, error) {
	var (
		values []interface{}
		err    error
	)
	if len(names) > 0 {
		values, err = cacheInstance.Client().HMGet(ctx, key, names...).Result()
	} else {
		valuesMap, getAllErr := cacheInstance.Client().HGetAll(ctx, key).Result()
		values = make([]interface{}, 0, len(valuesMap))
		for _, value := range valuesMap {
			values = append(values, value)
		}
		err = getAllErr
	}
	if err != nil {
		return nil, err
	}

	metrics := make([]*T, 0, len(values))
	if err := slices.UnmarshalBinary(values, &metrics); err != nil {
		return nil, err
	}
	return metrics, nil
}

func (c *cacheImpl) GetMetric(ctx context.Context, metricType vobj.MetricType, metricName string) (bo.MetricVec, error) {
	switch metricType {
	case vobj.MetricTypeCounter:
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeCounter)
		var metric bo.CounterMetricVec
		err := c.GetCache().Client().HGet(ctx, key, metricName).Scan(&metric)
		if err != nil {
			return nil, err
		}
		return &metric, nil
	case vobj.MetricTypeGauge:
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeGauge)
		var metric bo.GaugeMetricVec
		err := c.GetCache().Client().HGet(ctx, key, metricName).Scan(&metric)
		if err != nil {
			return nil, err
		}
		return &metric, nil
	case vobj.MetricTypeHistogram:
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeHistogram)
		var metric bo.HistogramMetricVec
		err := c.GetCache().Client().HGet(ctx, key, metricName).Scan(&metric)
		if err != nil {
			return nil, err
		}
		return &metric, nil
	case vobj.MetricTypeSummary:
		key := vobj.MetricCacheKeyPrefix.Key(vobj.MetricTypeSummary)
		var metric bo.SummaryMetricVec
		err := c.GetCache().Client().HGet(ctx, key, metricName).Scan(&metric)
		if err != nil {
			return nil, err
		}
		return &metric, nil
	default:
		return nil, merr.ErrorParams("invalid metric type: %s", metricType)
	}
}

func (c *cacheImpl) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.GetCache().Client().SetNX(ctx, key, 1, expiration).Result()
}

func (c *cacheImpl) Unlock(ctx context.Context, key string) error {
	return c.GetCache().Client().Del(ctx, key).Err()
}
