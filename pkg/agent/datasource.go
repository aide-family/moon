package agent

import (
	"context"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type (
	// Metric 查询到的数据labels
	Metric map[string]string

	// Result 查询到的数据
	Result struct {
		// Metric 查询到的数据labels
		Metric Metric `json:"metric"`
		// Ts 时间戳
		Ts float64 `json:"ts"`
		// Value 值
		Value float64 `json:"value"`
	}

	// Data 查询到的数据
	Data struct {
		// ResultType 查询到的数据类型
		ResultType string `json:"resultType"`
		// Result 查询到的数据集合
		Result []*Result `json:"result"`
	}

	// QueryResponse 查询结果
	QueryResponse struct {
		// Status 状态
		Status string `json:"status"`
		// Data 数据
		Data *Data `json:"data"`
		// ErrorType 错误类型
		ErrorType string `json:"errorType"`
		// Error 错误信息
		Error string `json:"error"`
	}

	// MetricDetail 查询到的数据详情， 用与元数据构建
	MetricDetail struct {
		// Name 指标名称
		Name string `json:"name"`
		// Help 帮助信息
		Help string `json:"help"`
		// Type 类型
		Type string `json:"type"`
		// Labels 标签集合
		Labels Labels `json:"labels"`
		// Unit 指标单位
		Unit string `json:"unit"`
	}

	// Metadata 查询到的元数据详情
	Metadata struct {
		// Metric 元数据列表
		Metric []*MetricDetail `json:"metric"`
		// Unix 查询时间戳
		Unix int64 `json:"unix"`
	}

	// Datasource 数据源完整接口定义
	Datasource interface {
		// Query 查询数据
		Query(ctx context.Context, expr string, duration int64) (*QueryResponse, error)
		// Metadata 查询元数据
		Metadata(ctx context.Context) (*Metadata, error)
		// GetCategory 获取数据源类型
		GetCategory() Category
		// GetEndpoint 获取数据源http地址
		GetEndpoint() string
		// GetBasicAuth 获取数据源http认证信息, 可选
		GetBasicAuth() *BasicAuth
		// WithBasicAuth 设置数据源http认证信息, 可选
		WithBasicAuth(basicAuth *BasicAuth) Datasource
	}

	// Category 数据源类型, 定义类型如下
	Category int32
)

const (
	// Prometheus Prometheus数据源类型，默认为该类型
	Prometheus Category = iota
	// VictoriaMetrics VictoriaMetrics数据源类型
	VictoriaMetrics
	// Elasticsearch Elasticsearch数据源类型
	Elasticsearch
	// Influxdb Influxdb数据源类型
	Influxdb
	// Clickhouse Clickhouse数据源类型
	Clickhouse
	// Loki Loki数据源类型
	Loki
)

// String Metric to string
func (m Metric) String() string {
	if m == nil {
		return ""
	}
	str := strings.Builder{}
	str.WriteString("{")
	keys := maps.Keys(m)
	// 排序
	sort.Strings(keys)
	for _, key := range keys {
		k := key
		v := m[key]
		str.WriteString(`"` + k + `"`)
		str.WriteString(":")
		str.WriteString(`"` + v + `"`)
		str.WriteString(",")
	}
	return strings.TrimRight(str.String(), ",") + "}"
}

// GetData get data
func (r *QueryResponse) GetData() *Data {
	if r == nil {
		return nil
	}
	return r.Data
}

// GetStatus get status
func (r *QueryResponse) GetStatus() string {
	if r == nil {
		return ""
	}
	return r.Status
}

// GetError get error
func (r *QueryResponse) GetError() string {
	if r == nil {
		return ""
	}
	return r.Error
}

// GetErrorType get error type
func (r *QueryResponse) GetErrorType() string {
	if r == nil {
		return ""
	}
	return r.ErrorType
}

// GetResultType get result type
func (d *Data) GetResultType() string {
	if d == nil {
		return ""
	}
	return d.ResultType
}

// GetResult get result
func (d *Data) GetResult() []*Result {
	if d == nil {
		return nil
	}
	return d.Result
}

// GetMetric get metric
func (r *Result) GetMetric() Metric {
	if r == nil {
		return nil
	}
	return r.Metric
}

// GetValue get value
func (r *Result) GetValue() float64 {
	if r == nil {
		return 0
	}
	return r.Value
}

// GetTs get ts
func (r *Result) GetTs() float64 {
	if r == nil {
		return 0
	}
	return r.Ts
}

var _category = map[Category]string{
	Prometheus:      "prometheus",
	VictoriaMetrics: "victoriametrics",
	Elasticsearch:   "elasticsearch",
	Influxdb:        "influxdb",
	Clickhouse:      "clickhouse",
	Loki:            "loki",
}

// String implements Stringer
func (c Category) String() string {
	remark, ok := _category[c]
	if !ok {
		return "unknown"
	}
	return remark
}

// Value return int32 value
func (c Category) Value() int32 {
	return int32(c)
}
