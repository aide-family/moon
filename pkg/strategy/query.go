package strategy

import (
	"context"
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
		ResultType string   `json:"resultType"`
		Result     []Result `json:"result"`
	}

	QueryResponse struct {
		Status string `json:"status"`
		Data   Data   `json:"data"`
	}
)

const (
	apiV1Query = "/api/v1/query"
)

const (
	metricName = "__name__"
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

// Bytes Metric to []byte
func (m Metric) Bytes() []byte {
	bs, _ := json.Marshal(m)
	return bs
}

// String Metric to string
func (m Metric) String() string {
	return string(m.Bytes())
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
