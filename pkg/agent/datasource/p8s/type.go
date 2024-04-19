package p8s

import (
	"github.com/aide-family/moon/pkg/agent/datasource"
)

type (
	Result struct {
		Metric datasource.Metric `json:"metric"`
		Value  []any             `json:"value"`
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
func (r *Result) GetMetric() datasource.Metric {
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
