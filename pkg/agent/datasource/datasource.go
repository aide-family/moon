package datasource

import (
	"context"
)

type (
	Metric map[string]string

	Result struct {
		Metric Metric  `json:"metric"`
		Ts     float64 `json:"ts"`
		Value  int64   `json:"value"`
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

	Datasource interface {
		Query(ctx context.Context, expr string, duration int64) (*QueryResponse, error)
		GetCategory() Category
		GetEndpoint() string
		GetBasicAuth() *BasicAuth
		WithBasicAuth(basicAuth *BasicAuth) Datasource
	}

	Category int32
)
