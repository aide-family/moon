package team

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.DatasourceMetric = (*DatasourceMetric)(nil)

const tableNameDatasourceMetric = "team_datasource_metrics"

type DatasourceMetric struct {
	do.TeamModel
	Name           string                        `gorm:"column:name;type:varchar(64);not null;comment:名称" json:"name"`
	Status         vobj.GlobalStatus             `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Remark         string                        `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Driver         vobj.DatasourceDriverMetric   `gorm:"column:type;type:tinyint(2);not null;comment:类型" json:"type"`
	Endpoint       string                        `gorm:"column:endpoint;type:varchar(255);not null;comment:数据源地址" json:"endpoint"`
	ScrapeInterval time.Duration                 `gorm:"column:scrape_interval;type:bigint(20);not null;comment:抓取间隔" json:"scrapeInterval"`
	Headers        *crypto.Object[[]*kv.KV]      `gorm:"column:headers;type:text;not null;comment:请求头" json:"headers"`
	QueryMethod    vobj.HTTPMethod               `gorm:"column:query_method;type:tinyint(2);not null;comment:请求方法" json:"queryMethod"`
	CA             crypto.String                 `gorm:"column:ca;type:text;not null;comment:ca" json:"ca"`
	TLS            *crypto.Object[*do.TLS]       `gorm:"column:tls;type:text;not null;comment:tls" json:"tls"`
	BasicAuth      *crypto.Object[*do.BasicAuth] `gorm:"column:basic_auth;type:text;not null;comment:basic_auth" json:"basicAuth"`
	Extra          *crypto.Object[[]*kv.KV]      `gorm:"column:extra;type:text;not null;comment:额外信息" json:"extra"`
	Metrics        []*StrategyMetric             `gorm:"many2many:team_strategy_metric_datasource" json:"metrics"`
}

func (d *DatasourceMetric) GetType() vobj.DatasourceType {
	return vobj.DatasourceTypeMetric
}

func (d *DatasourceMetric) GetStorageDriver() string {
	return d.Driver.String()
}

func (d *DatasourceMetric) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *DatasourceMetric) GetStatus() vobj.GlobalStatus {
	if d == nil {
		return vobj.GlobalStatusUnknown
	}
	return d.Status
}

func (d *DatasourceMetric) GetRemark() string {
	if d == nil {
		return ""
	}
	return d.Remark
}

func (d *DatasourceMetric) GetDriver() vobj.DatasourceDriverMetric {
	if d == nil {
		return vobj.DatasourceDriverMetricUnknown
	}
	return d.Driver
}

func (d *DatasourceMetric) GetEndpoint() string {
	if d == nil {
		return ""
	}
	return d.Endpoint
}

func (d *DatasourceMetric) GetScrapeInterval() time.Duration {
	if d == nil {
		return 0
	}
	return d.ScrapeInterval
}

func (d *DatasourceMetric) GetHeaders() []*kv.KV {
	if d == nil {
		return nil
	}
	return d.Headers.Get()
}

func (d *DatasourceMetric) GetQueryMethod() vobj.HTTPMethod {
	if d == nil {
		return vobj.HTTPMethodUnknown
	}
	return d.QueryMethod
}

func (d *DatasourceMetric) GetCA() string {
	if d == nil {
		return ""
	}
	return string(d.CA)
}

func (d *DatasourceMetric) GetTLS() *do.TLS {
	if d == nil {
		return nil
	}
	return d.TLS.Get()
}

func (d *DatasourceMetric) GetBasicAuth() *do.BasicAuth {
	if d == nil {
		return nil
	}
	return d.BasicAuth.Get()
}

func (d *DatasourceMetric) GetExtra() []*kv.KV {
	if d == nil {
		return nil
	}
	return d.Extra.Get()
}

func (d *DatasourceMetric) GetStrategies() []do.StrategyMetric {
	if d == nil {
		return nil
	}
	return slices.Map(d.Metrics, func(m *StrategyMetric) do.StrategyMetric { return m })
}

func (d *DatasourceMetric) TableName() string {
	return tableNameDatasourceMetric
}
