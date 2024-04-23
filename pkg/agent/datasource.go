package agent

import (
	"context"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
)

type (
	Metric map[string]string

	Result struct {
		Metric Metric  `json:"metric"`
		Ts     float64 `json:"ts"`
		Value  float64 `json:"value"`
	}

	Data struct {
		ResultType string    `json:"resultType"`
		Result     []*Result `json:"result"`
	}

	QueryResponse struct {
		Status    string `json:"status"`
		Data      *Data  `json:"data"`
		ErrorType string `json:"errorType"`
		Error     string `json:"error"`
	}

	MetricDetail struct {
		Name   string `json:"name"`
		Help   string `json:"help"`
		Type   string `json:"type"`
		Labels Labels `json:"labels"`
		Unit   string `json:"unit"`
	}

	Metadata struct {
		Metric []*MetricDetail `json:"metric"`
		Unix   int64           `json:"unix"`
	}

	Datasource interface {
		Query(ctx context.Context, expr string, duration int64) (*QueryResponse, error)
		Metadata(ctx context.Context) (*Metadata, error)
		GetCategory() Category
		GetEndpoint() string
		GetBasicAuth() *BasicAuth
		WithBasicAuth(basicAuth *BasicAuth) Datasource
	}

	Category int32
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
		str.WriteString(k)
		str.WriteString("=")
		str.WriteString(v)
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

const (
	Prometheus Category = iota
	VictoriaMetrics
	Elasticsearch
	Influxdb
	Clickhouse
)

var _category = map[Category]string{
	Prometheus:      "prometheus",
	VictoriaMetrics: "victoriametrics",
	Elasticsearch:   "elasticsearch",
	Influxdb:        "influxdb",
	Clickhouse:      "clickhouse",
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
