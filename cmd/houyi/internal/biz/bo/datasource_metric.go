package bo

import (
	"time"

	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/util/kv"
)

type MetricDatasourceConfig interface {
	cache.Object
	GetTeamId() uint32
	GetId() uint32
	GetName() string
	GetDriver() common.MetricDatasourceDriver
	GetEndpoint() string
	GetHeaders() []*kv.KV
	GetMethod() common.DatasourceQueryMethod
	GetBasicAuth() datasource.BasicAuth
	GetTLS() datasource.TLS
	GetCA() string
	GetEnable() bool
	GetScrapeInterval() time.Duration
}

type MetricQueryRequest struct {
	Expr string
	Time time.Time
}

type MetricRangeQueryRequest struct {
	Expr      string
	StartTime time.Time
	EndTime   time.Time
}

func (m *MetricRangeQueryRequest) GetOptimalStep(scrapeInterval time.Duration) time.Duration {
	duration := m.EndTime.Sub(m.StartTime)

	// Prometheus 通常会对较旧的数据进行降采样
	if duration > 15*24*time.Hour {
		// 对于超过15天的数据，使用较大的step
		return 2 * time.Hour
	} else if duration > 3*24*time.Hour {
		return 1 * time.Hour
	}

	// 确保step至少是scrape_interval的倍数
	minStep := scrapeInterval

	// 计算一个合理的step，使返回点数在500-1000之间
	desiredPoints := 800
	calculatedStep := duration / time.Duration(desiredPoints)

	// 确保step不小于最小step，且是scrapeInterval的倍数
	if calculatedStep < minStep {
		return minStep
	}

	// 向上取整到scrapeInterval的倍数
	return ((calculatedStep + scrapeInterval - 1) / scrapeInterval) * scrapeInterval
}

type SyncMetricMetadataRequest struct {
	Item       MetricDatasourceConfig
	OperatorId uint32
}

type MetricDatasourceQueryRequest struct {
	Datasource MetricDatasourceConfig
	Expr       string
	Time       int64
	StartTime  int64
	EndTime    int64
	Step       uint32
}

func (m *MetricDatasourceQueryRequest) IsQueryRange() bool {
	return m.EndTime > m.StartTime && m.EndTime > 0
}

func (m *MetricDatasourceQueryRequest) GetQueryRange() *MetricRangeQueryRequest {
	return &MetricRangeQueryRequest{
		Expr:      m.Expr,
		StartTime: time.Unix(m.StartTime, 0),
		EndTime:   time.Unix(m.EndTime, 0),
	}
}

func (m *MetricDatasourceQueryRequest) GetQuery() *MetricQueryRequest {
	return &MetricQueryRequest{
		Expr: m.Expr,
		Time: time.Unix(m.Time, 0),
	}
}

type MetricQueryValue struct {
	Value     float64
	Timestamp int64
}

type MetricQueryResult struct {
	Metric map[string]string
	Value  *MetricQueryValue
	Values []*MetricQueryValue
}

type MetricDatasourceQueryReply struct {
	Results []*MetricQueryResult
}
