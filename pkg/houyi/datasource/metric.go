package datasource

import (
	"context"

	"github.com/aide-family/moon/pkg/merr"
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
		Labels map[string]string `json:"labels"`
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

	// MetricDatasource 数据源完整接口定义
	MetricDatasource interface {
		Step() uint32
		// Query 查询数据
		Query(ctx context.Context, expr string, duration int64) ([]*QueryResponse, error)
		// QueryRange 查询数据
		QueryRange(ctx context.Context, expr string, start, end int64, step uint32) ([]*QueryResponse, error)
		// Metadata 查询元数据
		Metadata(ctx context.Context) (*Metadata, error)
		// GetBasicInfo 获取数据源信息
		GetBasicInfo() *BasicInfo
	}

	// BasicInfo 数据源信息
	BasicInfo struct {
		Endpoint  string     `json:"endpoint"`
		ID        uint32     `json:"id"`
		BasicAuth *BasicAuth `json:"basic_auth"`
	}

	datasourceBuild struct {
		endpoint  string
		id        uint32
		step      uint32
		basicAuth *BasicAuth
	}

	// MetricDatasourceBuildOption 数据源构建选项
	MetricDatasourceBuildOption func(p *datasourceBuild)
)

// NewMetricDatasource 创建数据源
func NewMetricDatasource(storageType vobj.StorageType, opts ...MetricDatasourceBuildOption) (MetricDatasource, error) {
	d := &datasourceBuild{}
	for _, opt := range opts {
		opt(d)
	}
	switch storageType {
	case vobj.StorageTypePrometheus:
		return NewPrometheusDatasource(
			WithPrometheusEndpoint(d.endpoint),
			WithPrometheusStep(d.step),
			WithPrometheusID(d.id),
			WithPrometheusBasicAuth(d.basicAuth.Username, d.basicAuth.Password),
		), nil
	case vobj.StorageTypeVictoriametrics:
		return NewVictoriaMetricsDatasource(
			WithVictoriaMetricsEndpoint(d.endpoint),
			WithVictoriaMetricsStep(d.step),
			WithVictoriaMetricsID(d.id),
			WithVictoriaMetricsBasicAuth(d.basicAuth.Username, d.basicAuth.Password),
		), nil
	default:
		return nil, merr.ErrorNotificationUnsupportedDataSource("unsupported data source type")
	}
}

// WithMetricEndpoint 设置数据源地址
func WithMetricEndpoint(endpoint string) MetricDatasourceBuildOption {
	return func(p *datasourceBuild) {
		p.endpoint = endpoint
	}
}

// WithMetricStep 设置数据源步长
func WithMetricStep(step uint32) MetricDatasourceBuildOption {
	return func(p *datasourceBuild) {
		p.step = step
	}
}

// WithMetricBasicAuth 设置数据源认证信息
func WithMetricBasicAuth(username, password string) MetricDatasourceBuildOption {
	return func(p *datasourceBuild) {
		p.basicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithMetricID 设置数据源ID
func WithMetricID(id uint32) MetricDatasourceBuildOption {
	return func(p *datasourceBuild) {
		p.id = id
	}
}

type mockMetricDatasource struct{}

func (m *mockMetricDatasource) GetBasicInfo() *BasicInfo {
	return new(BasicInfo)
}

func (m *mockMetricDatasource) Query(ctx context.Context, expr string, duration int64) ([]*QueryResponse, error) {
	return nil, nil
}

func (m *mockMetricDatasource) QueryRange(ctx context.Context, expr string, start, end int64, step uint32) ([]*QueryResponse, error) {
	return nil, nil
}

func (m *mockMetricDatasource) Metadata(ctx context.Context) (*Metadata, error) {
	return new(Metadata), nil
}

func (m *mockMetricDatasource) Step() uint32 {
	return 10
}

// NewMockMetricDatasource 创建一个mock metric数据源
func NewMockMetricDatasource() MetricDatasource {
	return &mockMetricDatasource{}
}
