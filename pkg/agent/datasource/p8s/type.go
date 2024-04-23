package p8s

import (
	"github.com/aide-family/moon/pkg/agent"
)

type (
	Result struct {
		Metric agent.Metric `json:"metric"`
		Value  []any        `json:"value"`
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
func (r *Result) GetMetric() agent.Metric {
	if r == nil {
		return nil
	}
	return r.Metric
}

// GetValue get value
func (r *Result) GetValue() []any {
	if r == nil {
		return nil
	}
	return r.Value
}

type (
	MetricInfo struct {
		Type string `json:"type"`
		Help string `json:"help"`
		Unit string `json:"unit"`
	}
	MetadataResponse struct {
		Status string                  `json:"status"`
		Data   map[string][]MetricInfo `json:"data"`
	}

	MetricLabel map[string]string

	MetricSeriesResponse struct {
		Status    string        `json:"status"`
		Data      []MetricLabel `json:"data"`
		Error     string        `json:"error"`
		ErrorType string        `json:"errorType"`
	}
)

func (m *MetadataResponse) GetStatus() string {
	if m == nil {
		return ""
	}
	return m.Status
}

func (m *MetadataResponse) GetData() map[string][]MetricInfo {
	if m == nil {
		return nil
	}
	return m.Data
}

func (m *MetadataResponse) GetMetricInfo(metric string) []MetricInfo {
	if m == nil {
		return nil
	}
	return m.Data[metric]
}

func (m *MetricInfo) GetType() string {
	if m == nil {
		return ""
	}
	return m.Type
}

func (m *MetricInfo) GetHelp() string {
	if m == nil {
		return ""
	}
	return m.Help
}

func (m *MetricInfo) GetUnit() string {
	if m == nil {
		return ""
	}
	return m.Unit
}

// GetMetricName get metric name
func (m MetricLabel) GetMetricName() string {
	if m == nil {
		return ""
	}
	return m["__name__"]
}

// GetData get data
func (m *MetricSeriesResponse) GetData() []MetricLabel {
	if m == nil {
		return nil
	}
	return m.Data
}
