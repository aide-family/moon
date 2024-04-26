package datasource

import (
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
)

type (
	datasourceBuilder struct {
		category agent.Category

		prometheusConfig []p8s.Option
	}

	Option func(*datasourceBuilder)
)

// getPrometheusConfig 获取Prometheus配置
func (d *datasourceBuilder) getPrometheusConfig() []p8s.Option {
	if pkg.IsNil(d) {
		return nil
	}
	return d.prometheusConfig
}

// getCategory 获取数据源类别
func (d *datasourceBuilder) getCategory() agent.Category {
	if pkg.IsNil(d) {
		return agent.Prometheus
	}

	return d.category
}
