package datasource

const (
	Prometheus Category = iota
	VictoriaMetrics
	Elasticsearch
	Influxdb
	Clickhouse
)

var _category = map[Category]string{
	Prometheus:      "prometheus",
	VictoriaMetrics: "victoriametrics",
	Elasticsearch:   "elasticsearch",
	Influxdb:        "influxdb",
	Clickhouse:      "clickhouse",
}

// String implements Stringer
func (c Category) String() string {
	return _category[c]
}
