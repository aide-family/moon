// Package datasource provides a client for datasource.
package datasource

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/aide-family/magicbox/httpx"
)

type MetricConfig interface {
	GetEndpoint() string
	GetHeaders() http.Header
	GetMethod() httpx.Method
	GetBasicAuth() *httpx.BasicAuth
	GetTLS() *tls.ConnectionState
	GetCA() string
	GetScrapeInterval() time.Duration
}

type MetricClient interface {
	Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, target string) error
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryRangeResponse, error)
	Query(ctx context.Context, query string, time time.Time) (*QueryResponse, error)
	Series(ctx context.Context, start, end time.Time, match []string) (*SeriesResponse, error)
	Metadata(ctx context.Context, metric string) (*MetadataResponse, error)
}
