package bo

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

type SaveTeamMetricDatasource struct {
	ID             uint32
	Name           string
	Status         vobj.GlobalStatus
	Remark         string
	Driver         vobj.DatasourceDriverMetric
	Endpoint       string
	ScrapeInterval time.Duration
	Headers        kv.StringMap
	QueryMethod    vobj.HTTPMethod
	CA             string
	TLS            *do.TLS
	BasicAuth      *do.BasicAuth
	Extra          kv.StringMap
}

type ListTeamMetricDatasource struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *ListTeamMetricDatasource) ToListTeamMetricDatasourceReply(datasourceItems []*team.DatasourceMetric) *ListTeamMetricDatasourceReply {
	return &ListTeamMetricDatasourceReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(datasourceItems, func(datasource *team.DatasourceMetric) do.DatasourceMetric { return datasource }),
	}
}

type ListTeamMetricDatasourceReply = ListReply[do.DatasourceMetric]

type UpdateTeamMetricDatasourceStatusRequest struct {
	DatasourceID uint32
	Status       vobj.GlobalStatus
}

type DatasourceMetricMetadata struct {
	Name         string
	Help         string
	Type         string
	Labels       map[string]string
	Unit         string
	DatasourceID uint32
}

type BatchSaveTeamMetricDatasourceMetadata struct {
	TeamID       uint32
	DatasourceID uint32
	Metadata     []*DatasourceMetricMetadata
	IsDone       bool
}

type SyncMetricMetadataRequest struct {
	TeamID       uint32
	DatasourceID uint32
}

type ListTeamMetricDatasourceMetadata struct {
	*PaginationRequest
	DatasourceID uint32
}

func (r *ListTeamMetricDatasourceMetadata) ToListTeamMetricDatasourceMetadataReply(metadataItems []*team.DatasourceMetricMetadata) *ListTeamMetricDatasourceMetadataReply {
	return &ListTeamMetricDatasourceMetadataReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(metadataItems, func(metadata *team.DatasourceMetricMetadata) do.DatasourceMetricMetadata { return metadata }),
	}
}

type ListTeamMetricDatasourceMetadataReply = ListReply[do.DatasourceMetricMetadata]

type UpdateTeamMetricDatasourceMetadataRemarkRequest struct {
	ID         uint32
	Help, Unit string
}

type MetricDatasourceQueryRequest struct {
	Datasource do.DatasourceMetric
	Expr       string
	Time       int64
	StartTime  int64
	EndTime    int64
	Step       uint32
}

var _ datasource.MetricConfig = (*metricDatasourceConfig)(nil)

func NewMetricDatasourceConfig(datasourceMetric do.DatasourceMetric) datasource.MetricConfig {
	return &metricDatasourceConfig{datasourceMetric: datasourceMetric}
}

type metricDatasourceConfig struct {
	datasourceMetric do.DatasourceMetric
}

// GetBasicAuth implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetBasicAuth() datasource.BasicAuth {
	return m.datasourceMetric.GetBasicAuth()
}

// GetCA implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetCA() string {
	return m.datasourceMetric.GetCA()
}

// GetEndpoint implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetEndpoint() string {
	return m.datasourceMetric.GetEndpoint()
}

// GetHeaders implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetHeaders() map[string]string {
	return m.datasourceMetric.GetHeaders()
}

// GetMethod implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetMethod() common.DatasourceQueryMethod {
	return common.DatasourceQueryMethod(m.datasourceMetric.GetQueryMethod().GetValue())
}

// GetScrapeInterval implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetScrapeInterval() time.Duration {
	return m.datasourceMetric.GetScrapeInterval()
}

// GetTLS implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetTLS() datasource.TLS {
	return m.datasourceMetric.GetTLS()
}
