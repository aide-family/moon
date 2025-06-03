package bo

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/common"
	houyicommon "github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/plugin/datasource/prometheus"
	"github.com/aide-family/moon/pkg/plugin/datasource/victoria"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type SaveTeamMetricDatasource struct {
	ID             uint32
	Name           string
	Remark         string
	Driver         vobj.DatasourceDriverMetric
	Endpoint       string
	ScrapeInterval int64
	Headers        []*kv.KV
	QueryMethod    vobj.HTTPMethod
	CA             string
	TLS            *do.TLS
	BasicAuth      *do.BasicAuth
	Extra          []*kv.KV
}

func (r *SaveTeamMetricDatasource) Validate() error {
	if r.ScrapeInterval <= 0 {
		return merr.ErrorParams("scrape interval must be greater than 0")
	}
	if !r.Driver.Exist() || r.Driver.IsUnknown() {
		return merr.ErrorParams("invalid datasource driver: %s", r.Driver)
	}
	if !r.QueryMethod.Exist() || r.QueryMethod.IsUnknown() {
		return merr.ErrorParams("invalid query method: %s", r.QueryMethod)
	}

	return nil
}

type ListTeamMetricDatasource struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *ListTeamMetricDatasource) ToListReply(datasourceItems []do.DatasourceMetric) *ListTeamMetricDatasourceReply {
	return &ListTeamMetricDatasourceReply{
		PaginationReply: r.ToReply(),
		Items:           datasourceItems,
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
	Labels       map[string][]string
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

type GetMetricDatasourceMetadataRequest struct {
	DatasourceID uint32
	ID           uint32
}

type ListTeamMetricDatasourceMetadata struct {
	*PaginationRequest
	DatasourceID uint32
	Keyword      string
	Type         string
}

func (r *ListTeamMetricDatasourceMetadata) ToListReply(metadataItems []do.DatasourceMetricMetadata) *ListTeamMetricDatasourceMetadataReply {
	return &ListTeamMetricDatasourceMetadataReply{
		PaginationReply: r.ToReply(),
		Items:           metadataItems,
	}
}

type ListTeamMetricDatasourceMetadataReply = ListReply[do.DatasourceMetricMetadata]

type UpdateMetricDatasourceMetadataRequest struct {
	DatasourceID uint32
	MetadataID   uint32
	Help         string
	Unit         string
	Type         string
}

type MetricDatasourceQueryRequest struct {
	Datasource do.DatasourceMetric
	Expr       string
	Time       int64
	StartTime  int64
	EndTime    int64
	Step       uint32
}

func (r *MetricDatasourceQueryRequest) IsQueryRange() bool {
	return r.EndTime >= r.StartTime && r.StartTime > 0
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
func (m *metricDatasourceConfig) GetHeaders() []*kv.KV {
	return m.datasourceMetric.GetHeaders()
}

// GetMethod implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetMethod() houyicommon.DatasourceQueryMethod {
	return houyicommon.DatasourceQueryMethod(m.datasourceMetric.GetQueryMethod().GetValue())
}

// GetScrapeInterval implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetScrapeInterval() time.Duration {
	return m.datasourceMetric.GetScrapeInterval()
}

// GetTLS implements datasource.MetricConfig.
func (m *metricDatasourceConfig) GetTLS() datasource.TLS {
	return m.datasourceMetric.GetTLS()
}

func ToMetricDatasource(datasourceMetric do.DatasourceMetric, logger log.Logger) (datasource.Metric, error) {
	config := NewMetricDatasourceConfig(datasourceMetric)
	switch datasourceMetric.GetDriver() {
	case vobj.DatasourceDriverMetricPrometheus:
		return prometheus.New(config, logger), nil
	case vobj.DatasourceDriverMetricVictoriametrics:
		return victoria.New(config, logger), nil
	default:
		return nil, merr.ErrorBadRequest("invalid datasource driver: %s", datasourceMetric.GetDriver())
	}
}

func ToMetricDatasourceQueryReply(reply *datasource.MetricQueryResponse, err error) (*common.MetricDatasourceQueryReply, error) {
	if err != nil {
		return nil, err
	}
	results := make([]*common.MetricQueryResult, 0, len(reply.Data.Result))
	for _, result := range reply.Data.Result {
		results = append(results, &common.MetricQueryResult{
			Metric: result.Metric,
			Value:  getMetricQueryResultValue(result.GetMetricQueryValue()),
			Values: getMetricQueryResultValues(result.GetMetricQueryValues()),
		})
	}
	return &common.MetricDatasourceQueryReply{
		Results: results,
	}, nil
}

func getMetricQueryResultValue(value *datasource.MetricQueryValue) *common.MetricQueryResultValue {
	if validate.IsNil(value) {
		return nil
	}
	return &common.MetricQueryResultValue{
		Timestamp: int64(value.Timestamp),
		Value:     value.Value,
	}
}

func getMetricQueryResultValues(values []*datasource.MetricQueryValue) []*common.MetricQueryResultValue {
	results := make([]*common.MetricQueryResultValue, 0, len(values))
	for _, value := range values {
		results = append(results, getMetricQueryResultValue(value))
	}
	return results
}

type DatasourceSelect struct {
	*PaginationRequest
	Keyword string
	Status  vobj.GlobalStatus
	Type    vobj.DatasourceType
}

func (r *DatasourceSelect) ToSelectReply(datasources []do.Datasource) *DatasourceSelectReply {
	return &DatasourceSelectReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(datasources, func(item do.Datasource) SelectItem {
			datasourceType := item.GetType().String()
			datasourceDriver := item.GetStorageDriver()
			return &selectItem{
				Value:    item.GetID(),
				Label:    item.GetName(),
				Disabled: !item.GetStatus().IsEnable() || item.GetDeletedAt() != 0,
				Extra: &selectItemExtra{
					Remark: item.GetRemark(),
					Icon:   datasourceDriver,
					Color:  datasourceType,
				},
			}
		}),
	}
}

type DatasourceSelectReply = ListReply[SelectItem]
