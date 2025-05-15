package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/moon-monitor/moon/pkg/api/houyi/common"
)

type ResultType string

type (
	MetricConfig interface {
		GetEndpoint() string
		GetHeaders() map[string]string
		GetMethod() common.DatasourceQueryMethod
		GetBasicAuth() BasicAuth
		GetTLS() TLS
		GetCA() string
		GetScrapeInterval() time.Duration
	}
	MetricQueryValue struct {
		Timestamp float64 `json:"timestamp"`
		Value     float64 `json:"value"`
	}

	MetricQueryResult struct {
		Metric map[string]string `json:"metric"`
		Value  []any             `json:"value"`
		Values [][]any           `json:"values"`
	}

	MetricQueryData struct {
		ResultType ResultType           `json:"resultType"`
		Result     []*MetricQueryResult `json:"result"`
	}

	MetricQueryResponse struct {
		Status    string           `json:"status"`
		Data      *MetricQueryData `json:"data"`
		ErrorType string           `json:"errorType"`
		Err       string           `json:"error"`
	}

	MetricQueryRequest struct {
		Expr      string
		Time      int64
		StartTime int64
		EndTime   int64
		Step      uint32
	}

	MetricMetadataItem struct {
		// Name metric name
		Name string `json:"name"`
		// Help metric help
		Help string `json:"help"`
		// Type metric type
		Type string `json:"type"`
		// Labels metric labels
		Labels map[string][]string `json:"labels"`
		// Unit metric unit
		Unit string `json:"unit"`
	}

	MetricMetadata struct {
		Metric    []*MetricMetadataItem `json:"metric"`
		Timestamp int64                 `json:"timestamp"`
	}
)

type Metric interface {
	Query(ctx context.Context, req *MetricQueryRequest) (*MetricQueryResponse, error)

	Metadata(ctx context.Context) (<-chan *MetricMetadata, error)

	GetScrapeInterval() time.Duration

	Proxy(ctx http.Context, target string) error
}

// IsSuccessResponse is response success
func (p *MetricQueryResponse) IsSuccessResponse() bool {
	return p.Status == "success"
}

// Error is response error
func (p *MetricQueryResponse) Error() string {
	return fmt.Sprintf("metric query failed: (%s) %s => %s", p.Status, p.ErrorType, p.Err)
}

// String json string
func (p *MetricQueryResponse) String() string {
	bs, _ := json.Marshal(p)
	return string(bs)
}

// GetMetricQueryValue get metric query value
func (m *MetricQueryResult) GetMetricQueryValue() *MetricQueryValue {
	if len(m.Values) > 0 || len(m.Value) != 2 {
		return nil
	}
	value := m.Value
	timestamp := value[0].(float64)
	val, _ := strconv.ParseFloat(value[1].(string), 64)
	return &MetricQueryValue{
		Timestamp: timestamp,
		Value:     val,
	}
}

func (m *MetricQueryResult) GetMetricQueryValues() []*MetricQueryValue {
	if len(m.Values) == 0 {
		return nil
	}
	list := make([]*MetricQueryValue, 0, len(m.Values))
	for _, v := range m.Values {
		if len(v) != 2 {
			continue
		}
		value := v
		timestamp := value[0].(float64)
		val, _ := strconv.ParseFloat(value[1].(string), 64)
		list = append(list, &MetricQueryValue{
			Timestamp: timestamp,
			Value:     val,
		})
	}
	return list
}

// String MetricQueryResult json string
func (m *MetricQueryResult) String() string {
	bs, _ := json.Marshal(m)
	return string(bs)
}
