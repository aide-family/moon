package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// vmMetricLookback 是指标名列表覆盖的时间窗口。VictoriaMetrics 的标签值接口
	// 在未指定 start 时只扫描最近 5 分钟，长期没有新样本的指标会漏掉。
	vmMetricLookback = 24 * time.Hour
	// vmMetricSeriesLimit 限制单个指标返回的序列数，避免详情请求拉回过大的响应。
	vmMetricSeriesLimit = 1000
	// maxMetricNameLength 是 Prometheus 兼容指标名的长度上限。
	maxMetricNameLength = 256
)

// vmMetricWindow 返回指标名和详情查询使用的时间窗口，统一按 Unix 秒表示。
func vmMetricWindow(now time.Time) (int64, int64) {
	end := now.Unix()
	return end - int64(vmMetricLookback/time.Second), end
}

func vmMetricNamesPath(start, end int64) string {
	values := url.Values{}
	values.Set("start", strconv.FormatInt(start, 10))
	values.Set("end", strconv.FormatInt(end, 10))
	return "/api/v1/label/__name__/values?" + values.Encode()
}

func vmMetricSeriesPath(name string, start, end int64) (string, error) {
	name = strings.TrimSpace(name)
	if !validMetricName(name) {
		return "", merr.Invalid("invalid request")
	}
	values := url.Values{}
	values.Set("match[]", `{__name__="`+name+`"}`)
	values.Set("start", strconv.FormatInt(start, 10))
	values.Set("end", strconv.FormatInt(end, 10))
	values.Set("limit", strconv.Itoa(vmMetricSeriesLimit))
	return "/api/v1/series?" + values.Encode(), nil
}

// listVictoriaMetricsMetrics 读取 VictoriaMetrics 的指标名。
//
// VictoriaMetrics 对 Prometheus 的 /api/v1/metadata 只返回空的兼容占位，
// 指标名要从 /api/v1/label/__name__/values 获取。该接口不提供类型、说明和单位，
// 这几列留空，由页面显示为占位符。
func (r *datasourceRepo) listVictoriaMetricsMetrics(ctx context.Context, item bo.Datasource) ([]bo.DatasourceMetric, error) {
	start, end := vmMetricWindow(time.Now())
	body, err := r.callUpstream(ctx, item, http.MethodGet, vmMetricNamesPath(start, end), "")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil || payload.Status != "success" {
		r.log.Error(fmt.Sprintf("datasource %s metric list", item.UID))
		return nil, merr.Client("upstream query failed")
	}
	seen := make(map[string]struct{}, len(payload.Data))
	out := make([]bo.DatasourceMetric, 0, len(payload.Data))
	for _, name := range payload.Data {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, bo.DatasourceMetric{Name: name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// victoriaMetricsMetricDetail 用 series 接口返回指标的所有标签组合。
// Prometheus 的 metadata 接口在 VictoriaMetrics 上没有内容，不能用来查详情。
func (r *datasourceRepo) victoriaMetricsMetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error) {
	start, end := vmMetricWindow(time.Now())
	path, err := vmMetricSeriesPath(name, start, end)
	if err != nil {
		return "", err
	}
	return r.callUpstream(ctx, item, http.MethodGet, path, "")
}

// validMetricName 校验 Prometheus 兼容的指标名：字母、数字、下划线和冒号，
// 首字符不能是数字。指标名会被拼进 PromQL 标签匹配器，不能放任意字符。
func validMetricName(name string) bool {
	if name == "" || len(name) > maxMetricNameLength {
		return false
	}
	for index, char := range name {
		switch {
		case char == '_', char == ':':
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case index > 0 && char >= '0' && char <= '9':
		default:
			return false
		}
	}
	return true
}
