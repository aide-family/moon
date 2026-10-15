package data

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

var upstreamClient = &http.Client{Timeout: 10 * time.Second}

func (r *repoBase) callUpstream(ctx context.Context, item bo.Datasource, method, path, body string) (string, error) {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return "", merr.Invalid("invalid request")
	}
	base, err := url.Parse(item.Endpoint)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") {
		return "", merr.Invalid("invalid request")
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", merr.Invalid("invalid request")
	}
	target := base.ResolveReference(ref)
	if target.Host != base.Host {
		return "", merr.Invalid("invalid request")
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), reader)
	if err != nil {
		return "", merr.Server("upstream query failed")
	}
	if item.Username != "" {
		req.SetBasicAuth(item.Username, item.Password)
	} else if item.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+item.Secret)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := upstreamClient.Do(req)
	if err != nil {
		r.log.Errorf("datasource %s upstream: %v", item.UID, err)
		return "", merr.Server("upstream query failed")
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		r.log.Errorf("datasource %s read: %v", item.UID, err)
		return "", merr.Server("upstream query failed")
	}
	if resp.StatusCode >= 300 {
		r.log.Errorf("datasource %s status %d", item.UID, resp.StatusCode)
		return "", merr.Client("upstream query failed")
	}
	return string(payload), nil
}

func (r *datasourceRepo) DatasourceStatus(ctx context.Context, item bo.Datasource) (string, error) {
	if _, err := r.callUpstream(ctx, item, http.MethodGet, "/api/v1/status/buildinfo", ""); err != nil {
		return "", err
	}
	return "up", nil
}

func (r *datasourceRepo) ListMetrics(ctx context.Context, item bo.Datasource) ([]string, error) {
	body, err := r.callUpstream(ctx, item, http.MethodGet, "/api/v1/label/__name__/values", "")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil || payload.Status != "success" {
		r.log.Errorf("datasource %s metric list", item.UID)
		return nil, merr.Client("upstream query failed")
	}
	return payload.Data, nil
}

func (r *datasourceRepo) MetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error) {
	path := "/api/v1/metadata?metric=" + url.QueryEscape(name)
	return r.callUpstream(ctx, item, http.MethodGet, path, "")
}

func (r *datasourceRepo) InstantQuery(ctx context.Context, item bo.Datasource, expr string, at int64) (string, error) {
	values := url.Values{}
	values.Set("query", expr)
	if at > 0 {
		values.Set("time", strconv.FormatInt(at, 10))
	}
	return r.callUpstream(ctx, item, http.MethodGet, "/api/v1/query?"+values.Encode(), "")
}

func (r *datasourceRepo) RangeQuery(ctx context.Context, item bo.Datasource, expr string, start, end, step int64) (string, error) {
	values := url.Values{}
	values.Set("query", expr)
	values.Set("start", strconv.FormatInt(start, 10))
	values.Set("end", strconv.FormatInt(end, 10))
	values.Set("step", strconv.FormatInt(step, 10))
	return r.callUpstream(ctx, item, http.MethodGet, "/api/v1/query_range?"+values.Encode(), "")
}

func (r *datasourceRepo) ProxyQuery(ctx context.Context, item bo.Datasource, method, path, body string) (string, error) {
	return r.callUpstream(ctx, item, method, path, body)
}
