package datasource

import (
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
)

func NewDataSource(opts ...Option) agent.Datasource {
	d := &datasourceBuilder{}
	for _, opt := range opts {
		opt(d)
	}
	return d.Builder()
}

func (d *datasourceBuilder) Builder() agent.Datasource {
	switch d.category {
	default:
		return p8s.NewPrometheusDatasource(d.getPrometheusConfig()...)
	}
}
