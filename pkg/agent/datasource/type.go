package datasource

import (
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
)

type (
	Config struct {
		Endpoint  string
		BasicAuth string
	}

	datasourceBuilder struct {
		category agent.DatasourceCategory

		prometheusConfig []p8s.Option
		config           *Config
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
func (d *datasourceBuilder) getCategory() agent.DatasourceCategory {
	if pkg.IsNil(d) {
		return agent.DatasourceCategoryPrometheus
	}

	return d.category
}

// getConfig 获取数据源配置
func (d *datasourceBuilder) getConfig() *Config {
	if pkg.IsNil(d) {
		return nil
	}

	return d.config
}

// getEndpoint 获取数据源Endpoint
func (d *Config) getEndpoint() string {
	if pkg.IsNil(d) {
		return ""
	}

	return d.Endpoint
}

// getBasicAuth 获取数据源BasicAuth
func (d *Config) getBasicAuth() *agent.BasicAuth {
	if pkg.IsNil(d) {
		return nil
	}

	return agent.NewBasicAuthWithString(d.BasicAuth)
}
