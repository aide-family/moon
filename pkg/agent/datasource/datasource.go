package datasource

import (
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/agent/datasource/p8s"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewDataSource(opts ...Option) (agent.Datasource, error) {
	d := &datasourceBuilder{}
	for _, opt := range opts {
		opt(d)
	}
	return d.Builder()
}

func (d *datasourceBuilder) Builder() (agent.Datasource, error) {
	switch d.category {
	case agent.DatasourceCategoryClickhouse:
		return nil, status.Error(codes.Unimplemented, "Clickhouse datasource not implemented")
	case agent.DatasourceCategoryVictoriaMetrics:
		return nil, status.Error(codes.Unimplemented, "VictoriaMetrics datasource not implemented")
	case agent.DatasourceCategoryElasticsearch:
		return nil, status.Error(codes.Unimplemented, "Elasticsearch datasource not implemented")
	case agent.DatasourceCategoryInfluxdb:
		return nil, status.Error(codes.Unimplemented, "Influxdb datasource not implemented")
	case agent.DatasourceCategoryLoki:
		return nil, status.Error(codes.Unimplemented, "Loki datasource not implemented")
	default:
		opts := d.getPrometheusConfig()
		opts = append(opts, p8s.WithEndpoint(d.getConfig().getEndpoint()), p8s.WithBasicAuth(d.getConfig().getBasicAuth()))
		return p8s.NewPrometheusDatasource(opts...), nil
	}
}
