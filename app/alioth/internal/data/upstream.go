package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

var upstreamClient = &http.Client{Timeout: 10 * time.Second}

// maxUpstreamBody 限制一次上游响应的读取量。日志查询可能返回上千行，上限比指标查询宽。
const maxUpstreamBody = 4 << 20

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
		r.log.Error(fmt.Sprintf("datasource %s upstream: %v", item.UID, err))
		return "", merr.Server("upstream query failed")
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamBody))
	if err != nil {
		r.log.Error(fmt.Sprintf("datasource %s read: %v", item.UID, err))
		return "", merr.Server("upstream query failed")
	}
	if resp.StatusCode >= 300 {
		r.log.Error(fmt.Sprintf("datasource %s status %d", item.UID, resp.StatusCode))
		return "", merr.Client("upstream query failed")
	}
	return string(payload), nil
}

func (r *datasourceRepo) DatasourceStatus(ctx context.Context, item bo.Datasource) (string, error) {
	path, err := statusPath(item.Driver)
	if err != nil {
		return "", err
	}
	if _, err := r.callUpstream(ctx, item, http.MethodGet, path, ""); err != nil {
		return "", err
	}
	return "up", nil
}

// statusPath 按驱动选择健康检查路径。Loki 用 /ready，Elasticsearch 用根路径，
// 指标数据源沿用 Prometheus 的 buildinfo。其他驱动没有约定的探测路径，
// 返回明确错误而不是探测一个无关接口。
func statusPath(driver string) (string, error) {
	switch driver {
	case "METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS":
		return "/api/v1/status/buildinfo", nil
	case "LOGS_LOKI":
		return "/ready", nil
	case "LOGS_ELASTICSEARCH":
		return "/", nil
	default:
		return "", merr.Invalid("unsupported datasource driver")
	}
}

func (r *datasourceRepo) ListMetrics(ctx context.Context, item bo.Datasource) ([]bo.DatasourceMetric, error) {
	body, err := r.callUpstream(ctx, item, http.MethodGet, "/api/v1/metadata", "")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status string `json:"status"`
		Data   map[string][]struct {
			Type string `json:"type"`
			Help string `json:"help"`
			Unit string `json:"unit"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil || payload.Status != "success" {
		r.log.Error(fmt.Sprintf("datasource %s metric list", item.UID))
		return nil, merr.Client("upstream query failed")
	}
	out := make([]bo.DatasourceMetric, 0, len(payload.Data))
	for name, entries := range payload.Data {
		metric := bo.DatasourceMetric{Name: name}
		if len(entries) > 0 {
			metric.Type = entries[0].Type
			metric.Help = entries[0].Help
			metric.Unit = entries[0].Unit
		}
		out = append(out, metric)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
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
