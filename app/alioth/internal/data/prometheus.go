package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// listPrometheusMetrics 读取 Prometheus 的指标元数据。
// 返回值包含类型、说明和单位，页面的元数据表直接展示这些字段。
func (r *datasourceRepo) listPrometheusMetrics(ctx context.Context, item bo.Datasource) ([]bo.DatasourceMetric, error) {
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

func (r *datasourceRepo) prometheusMetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error) {
	path := "/api/v1/metadata?metric=" + url.QueryEscape(name)
	return r.callUpstream(ctx, item, http.MethodGet, path, "")
}
