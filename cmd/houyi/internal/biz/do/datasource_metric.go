package do

import (
	"encoding/json"
	"time"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/datasource"
)

var _ cache.Object = (*DatasourceMetricConfig)(nil)

type DatasourceMetricConfig struct {
	TeamId         uint32                        `json:"teamId,omitempty"`
	ID             uint32                        `json:"id,omitempty"`
	Name           string                        `json:"name,omitempty"`
	Driver         common.MetricDatasourceDriver `json:"driver,omitempty"`
	Endpoint       string                        `json:"endpoint,omitempty"`
	Headers        map[string]string             `json:"headers,omitempty"`
	Method         common.DatasourceQueryMethod  `json:"method,omitempty"`
	CA             string                        `json:"ca,omitempty"`
	BasicAuth      *BasicAuth                    `json:"basicAuth,omitempty"`
	TLS            *TLS                          `json:"tls,omitempty"`
	Enable         bool                          `json:"enable,omitempty"`
	ScrapeInterval time.Duration                 `json:"scrapeInterval,omitempty"`
}

func (d *DatasourceMetricConfig) GetScrapeInterval() time.Duration {
	if d == nil || d.ScrapeInterval <= 0 {
		return 15 * time.Second
	}
	return d.ScrapeInterval
}

func (d *DatasourceMetricConfig) GetTeamId() uint32 {
	if d == nil {
		return 0
	}
	return d.TeamId
}

func (d *DatasourceMetricConfig) GetId() uint32 {
	if d == nil {
		return 0
	}
	return d.ID
}

func (d *DatasourceMetricConfig) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *DatasourceMetricConfig) GetEnable() bool {
	if d == nil {
		return false
	}
	return d.Enable
}

func (d *DatasourceMetricConfig) GetDriver() common.MetricDatasourceDriver {
	if d == nil {
		return common.MetricDatasourceDriver_METRIC_DATASOURCE_DRIVER_UNKNOWN
	}
	return d.Driver
}

func (d *DatasourceMetricConfig) GetEndpoint() string {
	if d == nil {
		return ""
	}
	return d.Endpoint
}

func (d *DatasourceMetricConfig) GetHeaders() map[string]string {
	if d == nil {
		return nil
	}
	return d.Headers
}

func (d *DatasourceMetricConfig) GetMethod() common.DatasourceQueryMethod {
	if d == nil {
		return common.DatasourceQueryMethod_POST
	}
	return d.Method
}

func (d *DatasourceMetricConfig) GetBasicAuth() datasource.BasicAuth {
	if d == nil {
		return nil
	}
	return d.BasicAuth
}

func (d *DatasourceMetricConfig) GetTLS() datasource.TLS {
	if d == nil {
		return nil
	}
	return d.TLS
}

func (d *DatasourceMetricConfig) GetCA() string {
	if d == nil {
		return ""
	}
	return d.CA
}

func (d *DatasourceMetricConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(d)
}

func (d *DatasourceMetricConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

func (d *DatasourceMetricConfig) UniqueKey() string {
	return vobj.MetricDatasourceUniqueKey(d.Driver, d.TeamId, d.ID)
}
