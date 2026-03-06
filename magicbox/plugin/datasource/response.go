package datasource

import (
	"fmt"
	"maps"
	"strconv"
)

type SeriesResponse struct {
	Status string                `json:"status"`
	Data   []*SeriesResponseData `json:"data"`
}

func (s *SeriesResponse) Error() string {
	if s.Status == "success" {
		return ""
	}
	return fmt.Sprintf("prometheus series response status: %s", s.Status)
}

type SeriesResponseData map[string]string

func (s *SeriesResponseData) Name() string {
	return (*s)["__name__"]
}

func (s *SeriesResponseData) Labels() map[string]string {
	m := maps.Clone(*s)
	delete(m, "__name__")
	return m
}

type MetadataResponse struct {
	Status string                `json:"status"`
	Data   *MetadataResponseData `json:"data"`
}

func (m *MetadataResponse) Error() string {
	if m.Status == "success" {
		return ""
	}
	return fmt.Sprintf("prometheus metadata response status: %s", m.Status)
}

type MetadataResponseData map[string][]*MetadataResponseDataItem

type MetadataResponseDataItem struct {
	Help string `json:"help"`
	Unit string `json:"unit"`
	Type string `json:"type"`
}

type QueryResponse struct {
	Status string             `json:"status"`
	Data   *QueryResponseData `json:"data"`
}

func (q *QueryResponse) Error() string {
	if q.Status == "success" {
		return ""
	}
	return fmt.Sprintf("prometheus query response status: %s", q.Status)
}

type ResultType string

const (
	ResultTypeMatrix ResultType = "matrix"
	ResultTypeVector ResultType = "vector"
)

type QueryResponseData struct {
	ResultType ResultType                     `json:"resultType"`
	Result     []*QueryResponseDataResultItem `json:"result"`
}

type QueryResponseDataResultItem struct {
	Metric SeriesResponseData `json:"metric"`
	Value  QueryResponseValue `json:"value"`
}

type QueryResponseValue []any

func (q QueryResponseValue) Timestamp() float64 {
	return toFloat64(q[0])
}

func (q QueryResponseValue) Value() float64 {
	return toFloat64(q[1])
}

// toFloat64 converts JSON-decoded value (float64 or string) to float64; Prometheus may return numbers as strings.
func toFloat64(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}

type QueryRangeResponse struct {
	Status string                  `json:"status"`
	Data   *QueryRangeResponseData `json:"data"`
}

func (q *QueryRangeResponse) Error() string {
	if q.Status == "success" {
		return ""
	}
	return fmt.Sprintf("prometheus query range response status: %s", q.Status)
}

type QueryRangeResponseData struct {
	ResultType ResultType                          `json:"resultType"`
	Result     []*QueryRangeResponseDataResultItem `json:"result"`
}

type QueryRangeResponseDataResultItem struct {
	Metric SeriesResponseData `json:"metric"`
	Values QueryRangeValues   `json:"values"`
}

type QueryRangeValues []QueryResponseValue
