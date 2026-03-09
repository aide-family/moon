// Package prometheus provides a client for Prometheus.
package prometheus

import (
	"github.com/aide-family/magicbox/plugin/datasource"
)

var _ datasource.MetricClient = (*Client)(nil)

func NewClient(c datasource.MetricConfig) *Client {
	return &Client{c: c, MetricClient: datasource.NewMetricClient(c)}
}

type Client struct {
	c datasource.MetricConfig
	datasource.MetricClient
}
