package data

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

// ListMetrics 按指标驱动读取指标名。Prometheus 读元数据接口；
// VictoriaMetrics 的元数据接口只是兼容占位，实际从标签值接口读取。
func (r *datasourceRepo) ListMetrics(ctx context.Context, item bo.Datasource) ([]bo.DatasourceMetric, error) {
	switch item.Driver {
	case driverMetricsPrometheus:
		return r.listPrometheusMetrics(ctx, item)
	case driverMetricsVictoriaMetrics:
		return r.listVictoriaMetricsMetrics(ctx, item)
	default:
		return nil, merr.Invalid("unsupported datasource driver")
	}
}

// MetricDetail 按指标驱动读取单个指标的详情。
func (r *datasourceRepo) MetricDetail(ctx context.Context, item bo.Datasource, name string) (string, error) {
	switch item.Driver {
	case driverMetricsPrometheus:
		return r.prometheusMetricDetail(ctx, item, name)
	case driverMetricsVictoriaMetrics:
		return r.victoriaMetricsMetricDetail(ctx, item, name)
	default:
		return "", merr.Invalid("unsupported datasource driver")
	}
}

func (r *datasourceRepo) InstantQuery(ctx context.Context, item bo.Datasource, expr string, at int64) (string, error) {
	if !isMetricsDriver(item.Driver) {
		return "", merr.Invalid("unsupported datasource driver")
	}
	values := url.Values{}
	values.Set("query", expr)
	if at > 0 {
		values.Set("time", strconv.FormatInt(at, 10))
	}
	return r.callUpstream(ctx, item, http.MethodGet, "/api/v1/query?"+values.Encode(), "")
}

func (r *datasourceRepo) RangeQuery(ctx context.Context, item bo.Datasource, expr string, start, end, step int64) (string, error) {
	path := "/api/v1/query_range"
	startValue := strconv.FormatInt(start, 10)
	endValue := strconv.FormatInt(end, 10)
	stepValue := strconv.FormatInt(step, 10)
	switch {
	case isMetricsDriver(item.Driver):
	case item.Driver == driverLogsLoki:
		path = "/loki/api/v1/query_range"
		startValue = strconv.FormatInt(time.Unix(start, 0).UnixNano(), 10)
		endValue = strconv.FormatInt(time.Unix(end, 0).UnixNano(), 10)
		stepValue += "s"
	default:
		return "", merr.Invalid("unsupported datasource driver")
	}
	values := url.Values{}
	values.Set("query", expr)
	values.Set("start", startValue)
	values.Set("end", endValue)
	values.Set("step", stepValue)
	return r.callUpstream(ctx, item, http.MethodGet, path+"?"+values.Encode(), "")
}

func (r *datasourceRepo) ProxyQuery(ctx context.Context, item bo.Datasource, method, path, body string) (string, error) {
	return r.callUpstream(ctx, item, method, path, body)
}
