// Package prometheus provides a client for Prometheus.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

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

func (p *Client) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) (*QueryRangeResponse, error) {
	hx := httpx.NewClient(httpx.GetHTTPClient())

	api, err := url.JoinPath(p.c.GetEndpoint(), "/api/v1/query_range")
	if err != nil {
		return nil, err
	}
	toURL, err := url.Parse(api)
	if err != nil {
		return nil, err
	}

	opts := []httpx.Option{
		httpx.WithHeaders(http.Header{
			"Accept":          {"*/*"},
			"Accept-Language": {"zh-CN,zh;q=0.9"},
			"Connection":      {"keep-alive"},
		}),
		httpx.WithBasicAuth(p.c.GetBasicAuth()),
		httpx.WithTLS(p.c.GetTLS()),
		httpx.WithQuery(url.Values{
			"query": {query},
			"start": {strconv.FormatInt(start.Unix(), 10)},
			"end":   {strconv.FormatInt(end.Unix(), 10)},
			"step":  {strconv.FormatInt(int64(step), 10)},
		}),
	}
	resp, err := hx.Do(ctx, httpx.MethodGet, toURL.String(), opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus query range response status: %s", resp.Status)
	}

	var queryRangeResponse QueryRangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryRangeResponse); err != nil {
		return nil, err
	}

	if err := queryRangeResponse.Error(); err != "" {
		return nil, &queryRangeResponse
	}
	return &queryRangeResponse, nil
}

func (p *Client) Query(ctx context.Context, query string, time time.Time) (*QueryResponse, error) {
	hx := httpx.NewClient(httpx.GetHTTPClient())

	api, err := url.JoinPath(p.c.GetEndpoint(), "/api/v1/query")
	if err != nil {
		return nil, err
	}
	toURL, err := url.Parse(api)
	if err != nil {
		return nil, err
	}

	opts := []httpx.Option{
		httpx.WithHeaders(http.Header{
			"Accept":          {"*/*"},
			"Accept-Language": {"zh-CN,zh;q=0.9"},
			"Connection":      {"keep-alive"},
		}),
		httpx.WithBasicAuth(p.c.GetBasicAuth()),
		httpx.WithTLS(p.c.GetTLS()),
		httpx.WithQuery(url.Values{
			"query": {query},
			"time":  {strconv.FormatInt(time.Unix(), 10)},
		}),
	}
	resp, err := hx.Do(ctx, httpx.MethodGet, toURL.String(), opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus query response status: %s", resp.Status)
	}

	var queryResponse QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResponse); err != nil {
		return nil, err
	}

	if err := queryResponse.Error(); err != "" {
		return nil, &queryResponse
	}
	return &queryResponse, nil
}

func (p *Client) Series(ctx context.Context, start, end time.Time, match []string) (*SeriesResponse, error) {
	hx := httpx.NewClient(httpx.GetHTTPClient())

	api, err := url.JoinPath(p.c.GetEndpoint(), "/api/v1/series")
	if err != nil {
		return nil, err
	}
	toURL, err := url.Parse(api)
	if err != nil {
		return nil, err
	}

	opts := []httpx.Option{
		httpx.WithHeaders(http.Header{
			"Accept":          {"*/*"},
			"Accept-Language": {"zh-CN,zh;q=0.9"},
			"Connection":      {"keep-alive"},
			"Content-Type":    {"application/x-www-form-urlencoded;charset=UTF-8"},
		}),
		httpx.WithBasicAuth(p.c.GetBasicAuth()),
		httpx.WithTLS(p.c.GetTLS()),
		httpx.WithBody([]byte(url.Values{
			"start":   {start.Format(time.RFC3339)},
			"end":     {end.Format(time.RFC3339)},
			"match[]": match,
		}.Encode())),
	}
	resp, err := hx.Do(ctx, httpx.MethodPost, toURL.String(), opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus series response status: %s", resp.Status)
	}

	var seriesResponse SeriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&seriesResponse); err != nil {
		return nil, err
	}

	if err := seriesResponse.Error(); err != "" {
		return nil, &seriesResponse
	}
	return &seriesResponse, nil
}

func (p *Client) Metadata(ctx context.Context, metric string) (*MetadataResponse, error) {
	hx := httpx.NewClient(httpx.GetHTTPClient())

	api, err := url.JoinPath(p.c.GetEndpoint(), "/api/v1/metadata")
	if err != nil {
		return nil, err
	}
	toURL, err := url.Parse(api)
	if err != nil {
		return nil, err
	}

	opts := []httpx.Option{
		httpx.WithHeaders(http.Header{
			"Accept":          {"*/*"},
			"Accept-Language": {"zh-CN,zh;q=0.9"},
			"Connection":      {"keep-alive"},
		}),
		httpx.WithBasicAuth(p.c.GetBasicAuth()),
		httpx.WithTLS(p.c.GetTLS()),
	}
	resp, err := hx.Do(ctx, httpx.MethodGet, toURL.String(), opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prometheus metadata response status: %s", resp.Status)
	}

	var metadataResponse MetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&metadataResponse); err != nil {
		return nil, err
	}

	if err := metadataResponse.Error(); err != "" {
		return nil, &metadataResponse
	}
	return &metadataResponse, nil
}
