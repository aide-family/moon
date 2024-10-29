package watch

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	watchQueueMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "watch_queue_size",
		Help: "The size of the watch queue",
	}, []string{"name"})

	watchStorageMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "watch_storage_size",
		Help: "The size of the watch storage",
	}, []string{"name"})
)

func init() {
	prometheus.MustRegister(watchQueueMetric)
	prometheus.MustRegister(watchStorageMetric)
}

func updateWatchQueueMetric(name string, size int) {
	watchQueueMetric.WithLabelValues(name).Set(float64(size))
}

func updateWatchStorageMetric(name string, size int) {
	watchStorageMetric.WithLabelValues(name).Set(float64(size))
}
