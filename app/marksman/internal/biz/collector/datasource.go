package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/magicbox/safety"
)

func NewDatasourceCollector() prometheus.Collector {
	return &datasourceCollector{
		collectors: safety.NewMap(make(map[string]prometheus.Collector)),
	}
}

type datasourceCollector struct {
	collectors *safety.Map[string, prometheus.Collector]
}

// Collect implements [prometheus.Collector].
func (d *datasourceCollector) Collect(ch chan<- prometheus.Metric) {
	d.collectors.Range(func(key string, value prometheus.Collector) bool {
		value.Collect(ch)
		return true
	})
}

// Describe implements [prometheus.Collector].
func (d *datasourceCollector) Describe(ch chan<- *prometheus.Desc) {
	d.collectors.Range(func(key string, value prometheus.Collector) bool {
		value.Describe(ch)
		return true
	})
}
