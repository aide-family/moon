package do

import (
	"time"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/kv"
)

type DatasourceMetric interface {
	TeamBase
	GetName() string
	GetStatus() vobj.GlobalStatus
	GetRemark() string
	GetDriver() vobj.DatasourceDriverMetric
	GetEndpoint() string
	GetScrapeInterval() time.Duration
	GetHeaders() kv.StringMap
	GetQueryMethod() vobj.HTTPMethod
	GetCA() string
	GetTLS() *TLS
	GetBasicAuth() *BasicAuth
	GetExtra() kv.StringMap
	GetStrategies() []StrategyMetric
}

type DatasourceMetricMetadata interface {
	TeamBase
	GetDatasourceMetricID() uint32
	GetDatasourceMetric() DatasourceMetric
	GetName() string
	GetHelp() string
	GetType() string
	GetLabels() map[string]string
	GetUnit() string
}
