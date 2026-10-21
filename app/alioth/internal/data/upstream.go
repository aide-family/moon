package data

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
	"github.com/aide-family/Alioth/pkg/upstream"
)

var upstreamClient = &http.Client{Timeout: 10 * time.Second}

// maxUpstreamBody 限制一次上游响应的读取量。日志查询可能返回上千行，上限比指标查询宽。
const maxUpstreamBody = 4 << 20

// maxUpstreamReason 限制错误响应摘要的长度。上游可能返回整页 HTML 或超长 JSON，
// 日志和调用方提示只需要第一段可读原因。
const maxUpstreamReason = 300

// upstreamReason 把上游错误响应压成一行摘要。
// VictoriaMetrics、Loki 等返回的 4xx 原因通常就是纯文本或 JSON 的 error 字段，
// 丢掉它只会留下无法定位的 “status 400”。摘要只用于日志和调用方提示，长度受限。
func upstreamReason(payload []byte) string {
	reason := strings.Join(strings.Fields(string(payload)), " ")
	if reason == "" {
		return ""
	}
	runes := []rune(reason)
	if len(runes) > maxUpstreamReason {
		return string(runes[:maxUpstreamReason]) + "…"
	}
	return reason
}

func (r *repoBase) callUpstream(ctx context.Context, item bo.Datasource, method, path, body string) (string, error) {
	target, err := upstream.Resolve(item.Endpoint, path)
	if err != nil {
		return "", err
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
		reason := upstreamReason(payload)
		if reason == "" {
			reason = http.StatusText(resp.StatusCode)
		}
		r.log.Error(fmt.Sprintf("datasource %s %s %s status %d: %s", item.UID, method, target.Path, resp.StatusCode, reason))
		return "", merr.Client("upstream query failed: " + reason)
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

// statusPath 按驱动选择健康检查路径。Loki 用 /ready，VictoriaLogs 用 /health，
// Elasticsearch 用根路径。Prometheus 和 VictoriaMetrics 都用 buildinfo：
// VictoriaMetrics 实现了这个兼容接口，且它相对数据源地址的路径前缀解析，
// 集群版 /select/<accountID>/prometheus 也能探测。
// 其他驱动没有约定的探测路径，返回明确错误而不是探测一个无关接口。
func statusPath(driver string) (string, error) {
	switch driver {
	case driverMetricsPrometheus, driverMetricsVictoriaMetrics:
		return "/api/v1/status/buildinfo", nil
	case driverLogsLoki:
		return "/ready", nil
	case driverLogsVlogs:
		return vlogsHealthPath, nil
	case driverLogsElasticsearch:
		return "/", nil
	default:
		return "", merr.Invalid("unsupported datasource driver")
	}
}
