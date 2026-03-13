package impl

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/plugin/datasource"
	"github.com/aide-family/magicbox/plugin/datasource/prometheus"
	"github.com/aide-family/magicbox/plugin/datasource/victoriametrics"

	"github.com/aide-family/marksman/internal/biz/bo"
)

const (
	metadataKeyBasicAuthUsername = "basicAuthUsername"
	metadataKeyBasicAuthPassword = "basicAuthPassword"
)

// datasourceMetricConfig adapts bo.DatasourceItemBo to datasource.MetricConfig for Prometheus/VM.
type datasourceMetricConfig struct {
	url      string
	metadata map[string]string
}

func (c *datasourceMetricConfig) GetEndpoint() string     { return c.url }
func (c *datasourceMetricConfig) GetHeaders() http.Header { return nil }
func (c *datasourceMetricConfig) GetMethod() httpx.Method { return httpx.MethodGet }
func (c *datasourceMetricConfig) GetBasicAuth() *httpx.BasicAuth {
	if c.metadata == nil {
		return nil
	}
	user, ok1 := c.metadata[metadataKeyBasicAuthUsername]
	pass, ok2 := c.metadata[metadataKeyBasicAuthPassword]
	if !ok1 || !ok2 || user == "" {
		return nil
	}
	return &httpx.BasicAuth{Username: user, Password: pass}
}
func (c *datasourceMetricConfig) GetTLS() *tls.ConnectionState     { return nil }
func (c *datasourceMetricConfig) GetCA() string                    { return "" }
func (c *datasourceMetricConfig) GetScrapeInterval() time.Duration { return 0 }

var _ datasource.MetricConfig = (*datasourceMetricConfig)(nil)

// NewMetricClientFromDatasource creates a MetricClient for the given metrics datasource.
// Returns (nil, merr.ErrorInvalidArgument) if the datasource is not a metrics type or driver is unsupported.
func NewMetricClientFromDatasource(ds *bo.DatasourceItemBo) (datasource.MetricClient, error) {
	if ds == nil || ds.Type != enum.DatasourceType_METRICS {
		return nil, merr.ErrorInvalidArgument("datasource is not a metrics type")
	}
	url := strings.TrimRight(ds.URL, "/")
	if url == "" {
		return nil, merr.ErrorInvalidArgument("datasource url is empty")
	}
	cfg := &datasourceMetricConfig{url: url, metadata: ds.Metadata}
	switch ds.Driver {
	case enum.DatasourceDriver_METRICS_PROMETHEUS:
		return prometheus.NewClient(cfg), nil
	case enum.DatasourceDriver_METRICS_VICTORIA_METRICS:
		return victoriametrics.NewClient(cfg), nil
	default:
		return nil, merr.ErrorInvalidArgument("unsupported datasource driver: %s", ds.Driver)
	}
}
