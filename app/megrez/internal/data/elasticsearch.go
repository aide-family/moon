package data

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aide-family/Alioth/app/megrez/internal/data/repository"
	"github.com/aide-family/Alioth/app/megrez/pkg/bo"
	"github.com/aide-family/Alioth/pkg/esagg"
	"github.com/aide-family/Alioth/pkg/merr"
	"github.com/aide-family/Alioth/pkg/upstream"
)

// esQuery 把 Elasticsearch 的检索表达式按时间桶计数，产出与 Prometheus matrix
// 等价的采样。规则表达式就是日志查询页面用的那份 Query DSL 或检索语句：
// 命中条数即采样值，阈值判定复用指标和 Loki 那一套。
type esQuery struct{}

func (esQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	if step <= 0 {
		step = esagg.DefaultStep
	}
	samples, err := esQuery{}.count(ctx, item, item.Expr, start, end, step)
	if err != nil {
		return nil, err
	}
	return samples, nil
}

// QueryInstant 为告警模板的 query 函数提供瞬时计数：取 at 之前的一个桶。
func (esQuery) QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error) {
	end := at.UTC().Unix()
	start := end - esagg.DefaultStep
	return esQuery{}.count(ctx, item, expr, start, end, esagg.DefaultStep)
}

// count 构造并执行一次 date_histogram 计数。返回单个序列：Elasticsearch 侧
// 一次查询只有一个计数口径，不像 Loki 那样按标签拆多条流。
func (esQuery) count(ctx context.Context, item bo.RuleSnapshot, expr string, start, end, step int64) ([]bo.Sample, error) {
	path, err := esagg.SearchPath(item.IndexPattern)
	if err != nil {
		// 索引模式在保存数据源时就校验过一次，这里只防御脏数据。
		return nil, merr.Invalid("invalid datasource index pattern")
	}
	target, err := upstream.Resolve(item.Endpoint, path)
	if err != nil {
		return nil, err
	}
	body, err := esagg.Body(esagg.Query{
		Expr: expr, IndexPattern: item.IndexPattern, TimeField: item.TimeField,
		Start: start, End: end, Step: step,
	})
	if err != nil {
		return nil, merr.Invalid("invalid query expression")
	}
	payload, err := requestDatasourceBody(ctx, item, http.MethodPost, target.String(), body)
	if err != nil {
		return nil, err
	}
	buckets, err := esagg.ParseBuckets(payload)
	if err != nil {
		// 响应里没有聚合结果说明上游不是预期的 Elasticsearch 响应，
		// 当成查询失败处理，避免把错误当成「零命中」去恢复事件。
		slog.Default().Error(fmt.Sprintf("rule %s elasticsearch: %v", item.UID, err))
		return nil, merr.Server("upstream query failed")
	}
	points := make([]bo.Point, 0, len(buckets))
	for _, bucket := range buckets {
		points = append(points, bo.Point{At: bucket.At, Value: bucket.Count, Raw: bucket.Raw})
	}
	return []bo.Sample{{Labels: map[string]string{}, Points: points, Step: step}}, nil
}

var _ repository.RangeQuery = esQuery{}
var _ repository.InstantQuery = esQuery{}
