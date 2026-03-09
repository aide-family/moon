// Package datasource provides a client for datasource.
package datasource

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

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
	prometheusv1.API
	// Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, target string) error
	// QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryRangeResponse, error)
	// Query(ctx context.Context, query string, time time.Time) (*QueryResponse, error)
	// Series(ctx context.Context, start, end time.Time, match []string) (*SeriesResponse, error)
	// Metadata(ctx context.Context, metric string) (*MetadataResponse, error)
	// // LabelNames returns all label names, optionally filtered by match[] selectors.
	// LabelNames(ctx context.Context, match []string) (*LabelNamesResponse, error)
	// // LabelValues returns all values for the given label name, optionally filtered by match[] selectors.
	// LabelValues(ctx context.Context, label string, match []string) (*LabelValuesResponse, error)
}

type metricClient struct {
	c MetricConfig
}

// Do implements [api.Client].
func (m *metricClient) Do(ctx context.Context, req *http.Request) (*http.Response, []byte, error) {
	client := httpx.GetHTTPClient()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

// URL implements [api.Client].
func (m *metricClient) URL(ep string, args map[string]string) *url.URL {
	u, err := url.Parse(m.c.GetEndpoint())
	if err != nil {
		return nil
	}
	u.Path = path.Join(u.Path, ep)
	q := u.Query()
	for k, v := range args {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	if m.c.GetBasicAuth() != nil {
		u.User = url.UserPassword(m.c.GetBasicAuth().Username, m.c.GetBasicAuth().Password)
	}
	return u
}

func NewMetricClient(c MetricConfig) MetricClient {
	return prometheusv1.NewAPI(&metricClient{c: c})
}
