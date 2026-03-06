// Package victoriametrics provides a client for VictoriaMetrics.
package victoriametrics

import (
	"context"
	"net/http"
	"time"

	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/plugin/datasource"
)

var _ datasource.MetricClient = (*Client)(nil)

func NewClient(c datasource.MetricConfig) *Client {
	return &Client{c: c}
}

type Client struct {
	c datasource.MetricConfig
}

// Metadata implements [datasource.MetricClient].
func (p *Client) Metadata(ctx context.Context, metric string) (*datasource.MetadataResponse, error) {
	panic("unimplemented")
}

// Query implements [datasource.MetricClient].
func (p *Client) Query(ctx context.Context, query string, time time.Time) (*datasource.QueryResponse, error) {
	panic("unimplemented")
}

// QueryRange implements [datasource.MetricClient].
func (p *Client) QueryRange(ctx context.Context, query string, start time.Time, end time.Time, step time.Duration) (*datasource.QueryRangeResponse, error) {
	panic("unimplemented")
}

// Series implements [datasource.MetricClient].
func (p *Client) Series(ctx context.Context, start time.Time, end time.Time, match []string) (*datasource.SeriesResponse, error) {
	panic("unimplemented")
}

func (p *Client) Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, target string) error {
	proxyClient := &httpx.ProxyClient{
		Host: p.c.GetEndpoint(),
	}
	return proxyClient.Proxy(ctx, w, r, target)
}
