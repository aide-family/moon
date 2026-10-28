package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/esagg"
	"github.com/aide-family/moon/pkg/merr"
)

// esDefaultTimeField 是 Elasticsearch 日志索引常用的时间字段。
// 数据源可以覆盖它，例如 @timestamp、event.time 或 timestamp。
const esDefaultTimeField = "@timestamp"

// queryElasticsearchLogs 把日志查询翻译成 Elasticsearch 的 _search 请求。
// query 既可以是 Query DSL 对象，也可以是普通检索语句；时间窗口和条数由用例层收敛。
func (r *datasourceRepo) queryElasticsearchLogs(ctx context.Context, item bo.Datasource, query string, start, end int64, limit int, direction string) (string, error) {
	clause, err := esQueryClause(query)
	if err != nil {
		return "", err
	}
	path, err := esSearchPath(item.IndexPattern)
	if err != nil {
		return "", err
	}
	timeField := strings.TrimSpace(item.TimeField)
	if timeField == "" {
		timeField = esDefaultTimeField
	}
	if len(timeField) > 128 {
		return "", merr.Invalid("invalid request")
	}
	order := "desc"
	if direction == "forward" {
		order = "asc"
	}
	payload := map[string]any{
		"size":             limit,
		"track_total_hits": true,
		"sort": []map[string]any{
			{timeField: map[string]any{"order": order}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"range": map[string]any{
							timeField: map[string]any{
								"gte": time.Unix(start, 0).UTC().Format(time.RFC3339Nano),
								"lte": time.Unix(end, 0).UTC().Format(time.RFC3339Nano),
							},
						},
					},
				},
				"must": []any{clause},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", merr.Server("upstream query failed")
	}
	return r.callUpstream(ctx, item, http.MethodPost, path, string(body))
}

// esQueryClause 兼容两种输入：完整的 Query DSL 对象，或者直接输入的检索语句。
// 前者原样作为 bool.must 的条件，后者包装成 query_string。
func esQueryClause(query string) (map[string]any, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, merr.Invalid("invalid request")
	}
	if !strings.HasPrefix(query, "{") {
		return map[string]any{"query_string": map[string]any{"query": query}}, nil
	}
	var clause map[string]any
	if err := json.Unmarshal([]byte(query), &clause); err != nil || len(clause) == 0 {
		return nil, merr.Invalid("invalid request")
	}
	return clause, nil
}

// esSearchPath 按索引模式选择搜索路径。索引模式为空时搜索全部索引。
// 索引模式只允许 Elasticsearch 索引名和通配符，避免把额外路径拼到上游。
func esSearchPath(indexPattern string) (string, error) {
	indexPattern = strings.TrimSpace(indexPattern)
	if indexPattern == "" {
		return "/_search", nil
	}
	if len(indexPattern) > 256 {
		return "", merr.Invalid("invalid request")
	}
	for _, char := range indexPattern {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '_', char == '.', char == '*', char == ',', char == '-':
		default:
			return "", merr.Invalid("invalid request")
		}
	}
	return "/" + indexPattern + "/_search", nil
}

// rangeQueryElasticsearch 执行告警详情图表的区间查询：把命中数按 step 切成时间桶，
// 再编码成 Prometheus matrix，前端和指标图表共用同一套解析。
func (r *datasourceRepo) rangeQueryElasticsearch(ctx context.Context, item bo.Datasource, expr string, start, end, step int64) (string, error) {
	if step <= 0 {
		step = esagg.DefaultStep
	}
	path, err := esagg.SearchPath(item.IndexPattern)
	if err != nil {
		return "", merr.Invalid("invalid request")
	}
	body, err := esagg.Body(esagg.Query{
		Expr: expr, IndexPattern: item.IndexPattern, TimeField: item.TimeField,
		Start: start, End: end, Step: step,
	})
	if err != nil {
		return "", merr.Invalid("invalid request")
	}
	payload, err := r.callUpstream(ctx, item, http.MethodPost, path, string(body))
	if err != nil {
		return "", err
	}
	buckets, err := esagg.ParseBuckets([]byte(payload))
	if err != nil {
		r.log.Error(fmt.Sprintf("datasource %s elasticsearch: %v", item.UID, err))
		return "", merr.Server("upstream query failed")
	}
	matrix, err := esagg.Matrix(buckets, nil)
	if err != nil {
		return "", merr.Server("upstream query failed")
	}
	return string(matrix), nil
}
