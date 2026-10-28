package data

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/app/megrez/internal/data/repository"
	"github.com/aide-family/moon/app/megrez/pkg/bo"
	"github.com/aide-family/moon/pkg/upstream"
)

const (
	lokiQueryRangePath = "/loki/api/v1/query_range"
	lokiQueryPath      = "/loki/api/v1/query"
)

type lokiQuery struct{}

// QueryRange 执行 Loki LogQL 指标表达式的区间查询。告警规则需返回数值矩阵；
// 原始日志流响应不会被解释为“没有命中”，而是作为查询错误拒绝。
func (lokiQuery) QueryRange(ctx context.Context, item bo.RuleSnapshot, start, end, step int64) ([]bo.Sample, error) {
	target, err := upstream.Resolve(item.Endpoint, lokiQueryRangePath)
	if err != nil {
		return nil, err
	}
	query := target.Query()
	query.Set("query", item.Expr)
	query.Set("start", strconv.FormatInt(time.Unix(start, 0).UnixNano(), 10))
	query.Set("end", strconv.FormatInt(time.Unix(end, 0).UnixNano(), 10))
	query.Set("step", strconv.FormatInt(step, 10)+"s")
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	return decodeMatrix(payload)
}

// QueryInstant 为 Loki LogQL 告警模板提供瞬时数值查询。
func (lokiQuery) QueryInstant(ctx context.Context, item bo.RuleSnapshot, expr string, at time.Time) ([]bo.Sample, error) {
	target, err := upstream.Resolve(item.Endpoint, lokiQueryPath)
	if err != nil {
		return nil, err
	}
	query := target.Query()
	query.Set("query", expr)
	query.Set("time", strconv.FormatInt(at.UnixNano(), 10))
	target.RawQuery = query.Encode()
	payload, err := requestDatasource(ctx, item, target.String())
	if err != nil {
		return nil, err
	}
	return decodeVector(payload)
}

var _ repository.RangeQuery = lokiQuery{}
var _ repository.InstantQuery = lokiQuery{}
