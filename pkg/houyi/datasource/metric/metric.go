package metric

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// BasicAuth 基础认证信息
	BasicAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// QueryValue 查询到的值
	QueryValue struct {
		Value     float64 `json:"value"`
		Timestamp int64   `json:"timestamp"`
	}

	// QueryResponse 查询到的响应
	QueryResponse struct {
		// 标签集合
		Labels *vobj.Labels `json:"labels"`
		// 值
		Value  *QueryValue   `json:"value"`
		Values []*QueryValue `json:"values"`
		// 结果类型
		ResultType string `json:"resultType"`
	}

	// Metric 查询到的数据详情， 用与元数据构建
	Metric struct {
		// Name 指标名称
		Name string `json:"name"`
		// Help 帮助信息
		Help string `json:"help"`
		// Type 类型
		Type string `json:"type"`
		// Labels 标签集合
		Labels map[string][]string `json:"labels"`
		// Unit 指标单位
		Unit string `json:"unit"`
	}

	// Metadata 查询到的元数据详情
	Metadata struct {
		// Metric 元数据列表
		Metric []*Metric `json:"metric"`
		// Timestamp 查询时间戳
		Timestamp int64 `json:"timestamp"`
	}

	// Datasource 数据源完整接口定义
	Datasource interface {
		// Query 查询数据
		Query(ctx context.Context, expr string, duration int64) ([]*QueryResponse, error)
		// QueryRange 查询数据
		QueryRange(ctx context.Context, expr string, start, end int64, step uint32) ([]*QueryResponse, error)
		// Metadata 查询元数据
		Metadata(ctx context.Context) (*Metadata, error)
	}

	datasourceBuild struct {
		prometheusOptions []PrometheusOption
	}

	// DatasourceBuildOption 数据源构建选项
	DatasourceBuildOption func(p *datasourceBuild)
)

// NewMetricDatasource 创建数据源
func NewMetricDatasource(storageType vobj.StorageType, opts ...DatasourceBuildOption) (Datasource, error) {
	d := &datasourceBuild{}
	for _, opt := range opts {
		opt(d)
	}
	switch storageType {
	case vobj.StorageTypePrometheus:
		return NewPrometheusDatasource(d.prometheusOptions...), nil
	default:
		return nil, merr.ErrorUnsupportedDatasourceTypeErr("unsupported data source type")
	}
}

// WithPrometheusOption 配置 prometheus 数据源
func WithPrometheusOption(opts ...PrometheusOption) DatasourceBuildOption {
	return func(p *datasourceBuild) {
		p.prometheusOptions = append(p.prometheusOptions, opts...)
	}
}
