package metric

import "github.com/prometheus/client_golang/prometheus"

var defaultLabels = []string{"kind", "operation", "code", "reason", "server"}

var RequestTotalMetric = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "moon",
		Subsystem: "request",
		Name:      "total",
		Help:      "The total number of requests",
	}, defaultLabels,
)

var RequestLatencyMetric = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "moon",
		Subsystem: "request",
		Name:      "latency_ms",
		Help:      "The latency of requests",
		Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
	}, defaultLabels,
)

func init() {
	prometheus.MustRegister(RequestTotalMetric, RequestLatencyMetric)
}
