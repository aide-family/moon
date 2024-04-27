package datasource

import (
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
)

// WithPrometheusConfig prometheus配置
func WithPrometheusConfig(prometheusConfig ...p8s.Option) Option {
	return func(builder *datasourceBuilder) {
		builder.prometheusConfig = append(builder.prometheusConfig, prometheusConfig...)
	}
}

// WithCategory 配置数据源类型
func WithCategory(category agent.DatasourceCategory) Option {
	return func(builder *datasourceBuilder) {
		builder.category = category
	}
}

// WithConfig 配置数据源
func WithConfig(config *Config) Option {
	return func(builder *datasourceBuilder) {
		if pkg.IsNil(config) {
			return
		}
		builder.config = config
	}
}
