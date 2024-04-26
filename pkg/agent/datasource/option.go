package datasource

import (
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
)

// WithPrometheusConfig prometheus配置
func WithPrometheusConfig(prometheusConfig ...p8s.Option) Option {
	return func(builder *datasourceBuilder) {
		builder.prometheusConfig = append(builder.prometheusConfig, prometheusConfig...)
	}
}
