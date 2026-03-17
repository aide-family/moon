package biz

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/repository"
)

// MetricQueryBiz handles metric datasource query, query_range and proxy.
type MetricQueryBiz struct {
	helper         *klog.Helper
	datasourceRepo repository.Datasource
	proxyRepo      repository.MetricDatasourceProxy
}

// NewMetricQuery creates the MetricQueryBiz.
func NewMetricQuery(
	datasourceRepo repository.Datasource,
	proxyRepo repository.MetricDatasourceProxy,
	helper *klog.Helper,
) *MetricQueryBiz {
	return &MetricQueryBiz{
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "metric_query")),
		datasourceRepo: datasourceRepo,
		proxyRepo:      proxyRepo,
	}
}

// Query runs an instant query and returns the raw JSON response (Prometheus /api/v1/query).
func (b *MetricQueryBiz) Query(ctx context.Context, uid snowflake.ID, query string, evalTime int64) (string, error) {
	ds, err := b.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return "", merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return "", merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	if evalTime <= 0 {
		evalTime = time.Now().Unix()
	}
	v := url.Values{}
	v.Set("query", query)
	v.Set("time", strconv.FormatInt(evalTime, 10))
	path := "api/v1/query?" + v.Encode()
	statusCode, body, err := b.proxyRepo.Proxy(ctx, ds, path, "GET", nil)
	if err != nil {
		b.helper.Errorw("msg", "metric query proxy failed", "error", err, "uid", uid)
		return "", err
	}
	if statusCode < 200 || statusCode >= 300 {
		b.helper.Errorw("msg", "metric query returned non-2xx", "status", statusCode, "uid", uid)
		return "", merr.ErrorInternalServer("datasource returned status %d", statusCode)
	}
	return string(body), nil
}

// QueryRange runs a range query and returns the raw JSON response (Prometheus /api/v1/query_range).
// If start, end or step are zero, defaults to last 1 hour: end=now, start=now-3600, step=60.
func (b *MetricQueryBiz) QueryRange(ctx context.Context, uid snowflake.ID, query string, start, end, step int64) (string, error) {
	ds, err := b.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return "", merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return "", merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	// Default to last 1 hour when not set.
	now := time.Now().Unix()
	if end <= 0 {
		end = now
	}
	if start <= 0 {
		start = end - 3600 // 1 hour ago
	}
	if step <= 0 {
		step = 60
	}
	v := url.Values{}
	v.Set("query", query)
	v.Set("start", strconv.FormatInt(start, 10))
	v.Set("end", strconv.FormatInt(end, 10))
	v.Set("step", strconv.FormatInt(step, 10))
	path := "api/v1/query_range?" + v.Encode()
	statusCode, body, err := b.proxyRepo.Proxy(ctx, ds, path, "GET", nil)
	if err != nil {
		b.helper.Errorw("msg", "metric query_range proxy failed", "error", err, "uid", uid)
		return "", err
	}
	if statusCode < 200 || statusCode >= 300 {
		b.helper.Errorw("msg", "metric query_range returned non-2xx", "status", statusCode, "uid", uid)
		return "", merr.ErrorInternalServer("datasource returned status %d", statusCode)
	}
	return string(body), nil
}

// Proxy forwards the request to the datasource and returns status code and body.
func (b *MetricQueryBiz) Proxy(ctx context.Context, uid snowflake.ID, path, method string, body []byte) (int, []byte, error) {
	ds, err := b.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return 0, nil, merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return 0, nil, merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	return b.proxyRepo.Proxy(ctx, ds, path, method, body)
}
