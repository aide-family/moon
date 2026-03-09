package impl

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/datasource"
	"github.com/aide-family/magicbox/plugin/datasource/prometheus"
	"github.com/aide-family/magicbox/plugin/datasource/victoriametrics"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
)

const (
	metricName            = "marksman_datasource_status"
	driverPrometheus      = "prometheus"
	driverVictoriaMetrics = "victoria_metrics"
)

// mainTsdbMetricConfig adapts conf.MainTsdb to datasource.MetricConfig (Prometheus/VM use same API).
type mainTsdbMetricConfig struct {
	url       string
	basicAuth *config.BasicAuthConfig
}

func (c *mainTsdbMetricConfig) GetEndpoint() string     { return c.url }
func (c *mainTsdbMetricConfig) GetHeaders() http.Header { return nil }
func (c *mainTsdbMetricConfig) GetMethod() httpx.Method { return httpx.MethodGet }
func (c *mainTsdbMetricConfig) GetBasicAuth() *httpx.BasicAuth {
	basicAuth := c.basicAuth
	if basicAuth == nil || !strings.EqualFold(basicAuth.GetEnabled(), "true") {
		return nil
	}
	return &httpx.BasicAuth{
		Username: basicAuth.GetUsername(),
		Password: basicAuth.GetPassword(),
	}
}
func (c *mainTsdbMetricConfig) GetTLS() *tls.ConnectionState     { return nil }
func (c *mainTsdbMetricConfig) GetCA() string                    { return "" }
func (c *mainTsdbMetricConfig) GetScrapeInterval() time.Duration { return 0 }

// NewMainTsdbQuerier returns a DatasourceStatusQuerier that uses the main TSDB (Prometheus or VictoriaMetrics).
// If mainTsdb config is missing or invalid, returns a no-op querier that returns empty result.
func NewMainTsdbQuerier(bc *conf.Bootstrap) repository.DatasourceStatusQuerier {
	cfg := bc.GetMainTsdb()
	if cfg == nil || strings.TrimSpace(cfg.GetUrl()) == "" {
		return &noopStatusQuerier{}
	}
	driver := cfg.GetDriver()
	switch driver {
	case enum.DatasourceDriver_METRICS_PROMETHEUS:
		adapter := &mainTsdbMetricConfig{url: strings.TrimRight(cfg.GetUrl(), "/"), basicAuth: cfg.GetBasicAuth()}
		client := prometheus.NewClient(adapter)
		return &mainTsdbQuerier{client: client}
	case enum.DatasourceDriver_METRICS_VICTORIA_METRICS:
		adapter := &mainTsdbMetricConfig{url: strings.TrimRight(cfg.GetUrl(), "/"), basicAuth: cfg.GetBasicAuth()}
		client := victoriametrics.NewClient(adapter)
		return &mainTsdbQuerier{client: client}
	default:
		return &noopStatusQuerier{}
	}
}

type noopStatusQuerier struct{}

func (n *noopStatusQuerier) QueryDatasourceStatus(_ context.Context, _ *bo.GetDatasourceStatusRequest) ([]*bo.DatasourceStatusSeriesBo, error) {
	return nil, nil
}

type mainTsdbQuerier struct {
	client datasource.MetricClient
}

func (q *mainTsdbQuerier) QueryDatasourceStatus(ctx context.Context, req *bo.GetDatasourceStatusRequest) ([]*bo.DatasourceStatusSeriesBo, error) {
	// Prometheus expects label values in double quotes; Step is already time.Duration from GetStep().
	query := fmt.Sprintf(`%s{uid="%s",name="%s"}`, metricName, strconv.FormatInt(req.GetUID(), 10), req.GetName())
	queryRange := prometheusv1.Range{
		Start: time.Unix(req.GetStartTime(), 0),
		End:   time.Unix(req.GetEndTime(), 0),
		Step:  req.GetStep(),
	}
	value, _, err := q.client.QueryRange(ctx, query, queryRange)
	if err != nil {
		return nil, merr.ErrorInternalServer("query main tsdb failed").WithCause(err)
	}

	matrix, ok := value.(model.Matrix)
	if !ok || len(matrix) == 0 {
		return []*bo.DatasourceStatusSeriesBo{}, nil
	}

	out := make([]*bo.DatasourceStatusSeriesBo, 0, len(matrix))
	for _, stream := range matrix {
		points := make([]bo.DatasourceStatusPointBo, 0, len(stream.Values))
		for _, p := range stream.Values {
			points = append(points, bo.DatasourceStatusPointBo{
				Timestamp: int64(p.Timestamp) / 1000, // Prometheus model.Time is milliseconds
				Value:     float64(p.Value),
			})
		}
		out = append(out, &bo.DatasourceStatusSeriesBo{
			UID:    req.GetUID(),
			Name:   req.GetName(),
			Points: points,
		})
	}
	return out, nil
}

var _ datasource.MetricConfig = (*mainTsdbMetricConfig)(nil)
