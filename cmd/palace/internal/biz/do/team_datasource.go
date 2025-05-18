package do

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/kv"
)

type Datasource interface {
	TeamBase
	GetName() string
	GetRemark() string
	GetType() vobj.DatasourceType
	GetStorageDriver() string
	GetStatus() vobj.GlobalStatus
}

type DatasourceMetric interface {
	Datasource
	GetDriver() vobj.DatasourceDriverMetric
	GetEndpoint() string
	GetScrapeInterval() time.Duration
	GetHeaders() []*kv.KV
	GetQueryMethod() vobj.HTTPMethod
	GetCA() string
	GetTLS() *TLS
	GetBasicAuth() *BasicAuth
	GetExtra() []*kv.KV
	GetStrategies() []StrategyMetric
}

type DatasourceMetricMetadata interface {
	TeamBase
	GetDatasourceMetricID() uint32
	GetDatasourceMetric() DatasourceMetric
	GetName() string
	GetHelp() string
	GetType() string
	GetLabels() map[string][]string
	GetUnit() string
}
