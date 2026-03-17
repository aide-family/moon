package repository

import (
	"context"

	"github.com/aide-family/marksman/internal/biz/bo"
)

// MetricDatasourceProxy forwards HTTP requests to a metric-type datasource and returns the response.
type MetricDatasourceProxy interface {
	// Proxy sends an HTTP request to the datasource at the given path and returns status code and body.
	// Path is the suffix without leading slash (e.g. "api/v1/query", "api/v1/query_range?query=up").
	Proxy(ctx context.Context, ds *bo.DatasourceItemBo, path, method string, body []byte) (statusCode int, responseBody []byte, err error)
}
