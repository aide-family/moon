// Package datasource provides a client for datasource.
package datasource

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
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

// URL implements [api.Client]. Path args (e.g. :name in "/label/:name/values") must be
// replaced in the path; the Prometheus client passes them in args and expects substitution.
func (m *metricClient) URL(ep string, args map[string]string) *url.URL {
	u, err := url.Parse(m.c.GetEndpoint())
	if err != nil {
		return nil
	}
	p := path.Join(u.Path, ep)
	for k, v := range args {
		p = strings.ReplaceAll(p, ":"+k, v)
	}
	if p != "" && p[0] != '/' {
		p = "/" + p
	}
	u.Path = p
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
