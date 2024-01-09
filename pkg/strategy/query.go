package strategy

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
)

type Query interface {
	Query(ctx context.Context, expr string, duration int64) (*QueryResponse, error)
}

type (
	Metric map[string]string

	Result struct {
		Metric Metric `json:"metric"`
		Value  []any  `json:"value"`
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
)

const (
	apiV1Query = "/api/v1/query"
)

const (
	metricName            = "__name__"
	metricGroupName       = "__group_name__"
	metricAlert           = "__alert__"
	metricRuleLabelPrefix = "__rule_label__"
)

// Bytes QueryResponse to []byte
func (qr *QueryResponse) Bytes() []byte {
	bs, _ := json.Marshal(qr)
	return bs
}

// String QueryResponse to string
func (qr *QueryResponse) String() string {
	return string(qr.Bytes())
}

// Name Metric __name__
func (m Metric) Name() string {
	return m.Get(metricName)
}

// Get get tag value
func (m Metric) Get(key string) string {
	return m[key]
}

// Set Metric set tag value
func (m Metric) Set(key, value string) {
	m[key] = value
}

// Bytes Metric to []byte
func (m Metric) Bytes() []byte {
	if m == nil {
		return nil
	}
	bs, _ := json.Marshal(m)
	return bs
}

// String Metric to string
func (m Metric) String() string {
	return string(m.Bytes())
}

// MD5 Metric to md5
func (m Metric) MD5() string {
	return fmt.Sprintf("%x", md5.Sum(m.Bytes()))
}

// GetMetric Result to Metric
func (r *Result) GetMetric() Metric {
	if r == nil {
		return nil
	}
	return r.Metric
}

// ParseQuery 处理结构体转为query参数
func ParseQuery(qr map[string]any) (string, error) {
	if len(qr) == 0 {
		return "", fmt.Errorf("query is empty")
	}
	query := url.Values{}
	for k, v := range qr {
		query.Add(k, fmt.Sprintf("%v", v))
	}
	return query.Encode(), nil
}
