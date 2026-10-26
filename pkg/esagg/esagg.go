// Package esagg 把 Elasticsearch 的日志检索表达式翻译成「按时间桶的文档计数」。
//
// 告警评测需要的是每个时间桶一个数值（Prometheus matrix 的语义），而 Elasticsearch
// 的 _search 返回的是命中文档。这里用 date_histogram 把命中数按 step 切桶，
// 让 Elasticsearch 也能参与 Loki / VictoriaLogs 那一套阈值判定。
// 评测（megrez）和事件图表（alioth）共用同一份请求构造与响应解析。
package esagg

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultTimeField 是日志索引最常用的时间字段，数据源可以覆盖。
	DefaultTimeField = "@timestamp"
	// DefaultStep 是调用方没有给出桶宽时使用的秒数，与评测调度周期一致。
	DefaultStep = 15
	// maxIndexPatternLength 限制索引模式长度，避免把超长路径拼到上游。
	maxIndexPatternLength = 256
	// maxTimeFieldLength 限制时间字段名长度。
	maxTimeFieldLength = 128
	// maxExprLength 限制表达式长度。
	maxExprLength = 4096
	// aggName 是请求里聚合的名字，解析时按同一个名字取值。
	aggName = "buckets"
)

// Query 描述一次按时间桶计数的查询。
//
// Start 含、End 不含，两者都是 Unix 秒；调用方负责把 End 对齐到 Step 的整数倍，
// 否则 date_histogram 的桶边界和查询区间会错位，最后一个桶只统计了部分文档。
type Query struct {
	// Expr 是 Query DSL 对象或普通检索语句，与日志查询页面的输入格式一致。
	Expr string
	// IndexPattern 是索引模式，空表示搜索全部索引。
	IndexPattern string
	// TimeField 是时间字段，空表示 DefaultTimeField。
	TimeField string
	// Start 是查询区间左端，Unix 秒，含。
	Start int64
	// End 是查询区间右端，Unix 秒，不含。
	End int64
	// Step 是桶宽，秒。
	Step int64
}

// Bucket 是一个时间桶的文档数。
type Bucket struct {
	// At 是桶的起始时间。
	At time.Time
	// Count 是桶内的文档数。
	Count float64
	// Raw 是原始计数文本，用于告警事件的采样值展示。
	Raw string
}

// TimeField 返回实际使用的时间字段。
func (q Query) TimeFieldOrDefault() string {
	field := strings.TrimSpace(q.TimeField)
	if field == "" {
		return DefaultTimeField
	}
	return field
}

// SearchPath 按索引模式选择搜索路径。索引模式只允许索引名字符和通配符，
// 避免把额外路径拼到上游。
func SearchPath(indexPattern string) (string, error) {
	pattern := strings.TrimSpace(indexPattern)
	if pattern == "" {
		return "/_search", nil
	}
	if len(pattern) > maxIndexPatternLength {
		return "", fmt.Errorf("index pattern is too long")
	}
	for _, char := range pattern {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '_', char == '.', char == '*', char == ',', char == '-':
		default:
			return "", fmt.Errorf("invalid index pattern")
		}
	}
	return "/" + pattern + "/_search", nil
}

// Body 构造按时间桶计数的 _search 请求体。
func Body(q Query) ([]byte, error) {
	clause, err := clause(q.Expr)
	if err != nil {
		return nil, err
	}
	step := q.Step
	if step <= 0 {
		step = DefaultStep
	}
	if q.End <= q.Start {
		return nil, fmt.Errorf("invalid query window")
	}
	field := q.TimeFieldOrDefault()
	if len(field) > maxTimeFieldLength || strings.ContainsAny(field, " \t\r\n\"'{}[]") {
		return nil, fmt.Errorf("invalid time field")
	}
	startMs := q.Start * 1000
	endMs := q.End * 1000
	payload := map[string]any{
		// 只要聚合结果，不取命中文档。
		"size":             0,
		"track_total_hits": false,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{
						"range": map[string]any{
							// 半开区间 [start, end)，和桶的划分保持一致。
							field: map[string]any{"gte": startMs, "lt": endMs},
						},
					},
				},
				"must": []any{clause},
			},
		},
		"aggs": map[string]any{
			aggName: map[string]any{
				"date_histogram": map[string]any{
					"field":          field,
					// empty 桶也返回，语义与 Loki 的 count_over_time 一致：
					// FOR 条件需要看到「这段时间一直是 0」而不是「没有点」。
					"min_doc_count":   0,
					"fixed_interval":  strconv.FormatInt(step, 10) + "s",
					"extended_bounds": map[string]any{"min": startMs, "max": endMs - 1},
				},
			},
		},
	}
	return json.Marshal(payload)
}

// ParseBuckets 解析 date_histogram 的桶。响应里没有聚合结果时返回错误，
// 而不是当成「零命中」——查询失败不能被误判成告警恢复。
func ParseBuckets(payload []byte) ([]Bucket, error) {
	var body struct {
		Aggregations map[string]struct {
			Buckets []struct {
				Key      *json.Number `json:"key"`
				DocCount json.Number  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"aggregations"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&body); err != nil {
		return nil, fmt.Errorf("decode elasticsearch response: %w", err)
	}
	agg, ok := body.Aggregations[aggName]
	if !ok {
		return nil, fmt.Errorf("elasticsearch response has no %q aggregation", aggName)
	}
	buckets := make([]Bucket, 0, len(agg.Buckets))
	for _, item := range agg.Buckets {
		if item.Key == nil {
			return nil, fmt.Errorf("elasticsearch bucket has no key")
		}
		key, err := item.Key.Int64()
		if err != nil {
			return nil, fmt.Errorf("decode elasticsearch bucket key: %w", err)
		}
		count, err := item.DocCount.Float64()
		if err != nil {
			return nil, fmt.Errorf("decode elasticsearch doc_count: %w", err)
		}
		buckets = append(buckets, Bucket{
			At:    time.UnixMilli(key).UTC(),
			Count: count,
			Raw:   item.DocCount.String(),
		})
	}
	return buckets, nil
}

// Matrix 把桶序列编码成 Prometheus matrix，供事件图表直接使用，
// 与 query_range 返回的 JSON 同形。
func Matrix(buckets []Bucket, labels map[string]string) ([]byte, error) {
	if labels == nil {
		labels = map[string]string{}
	}
	values := make([][]any, 0, len(buckets))
	for _, bucket := range buckets {
		values = append(values, []any{float64(bucket.At.Unix()), strconv.FormatFloat(bucket.Count, 'f', -1, 64)})
	}
	body := map[string]any{
		"status": "success",
		"data": map[string]any{
			"resultType": "matrix",
			"result": []any{
				map[string]any{"metric": labels, "values": values},
			},
		},
	}
	return json.Marshal(body)
}

// clause 兼容两种输入：完整的 Query DSL 对象，或者直接输入的检索语句。
// 前者原样作为 bool.must 的条件，后者包装成 query_string。
func clause(query string) (map[string]any, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query expression")
	}
	if len(query) > maxExprLength {
		return nil, fmt.Errorf("query expression is too long")
	}
	if !strings.HasPrefix(query, "{") {
		return map[string]any{"query_string": map[string]any{"query": query}}, nil
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(query), &parsed); err != nil || len(parsed) == 0 {
		return nil, fmt.Errorf("invalid query dsl")
	}
	return parsed, nil
}
