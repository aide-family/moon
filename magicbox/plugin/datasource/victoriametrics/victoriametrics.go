// Package victoriametrics provides a client for VictoriaMetrics.
package victoriametrics

import (
	"context"
	"net/http"

	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/plugin/datasource"
)

type Client struct {
	c datasource.MetricConfig
}

func NewClient(c datasource.MetricConfig) *Client {
	return &Client{c: c}
}

func (p *Client) Proxy(ctx context.Context, w http.ResponseWriter, r *http.Request, target string) error {
	proxyClient := &httpx.ProxyClient{
		Host: p.c.GetEndpoint(),
	}
	return proxyClient.Proxy(ctx, w, r, target)
}
