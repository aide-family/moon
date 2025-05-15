package vobj

// DatasourceDriverMetric represents the metric type of a datasource driver.
//
//go:generate stringer -type=DatasourceDriverMetric -linecomment -output=datasource_driver_metric.string.go
type DatasourceDriverMetric int8

const (
	DatasourceDriverMetricUnknown         DatasourceDriverMetric = iota // unknown
	DatasourceDriverMetricPrometheus                                    // prometheus
	DatasourceDriverMetricVictoriametrics                               // victoriametrics
)
