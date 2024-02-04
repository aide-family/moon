package prom

import "github.com/prometheus/client_golang/prometheus"

var (
	MetricSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "server requests duration(µs).",
		Buckets:   []float64{100, 250, 500, 1000, 2500, 5000, 10000, 20000},
	}, []string{"kind", "operation"})

	MetricRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "client",
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "The total number of processed requests",
	}, []string{"kind", "operation", "code", "reason"})

	IpMetricCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "ip_total",
		Help:      "The total number of processed requests",
	}, []string{"ip"})
)

func init() {
	prometheus.MustRegister(MetricSeconds, MetricRequests, IpMetricCounter)
}
