package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var requestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "request_total",
		Help:      "count of requests received",
	},
	[]string{"component", "path"},
)

func init() {
	prometheus.MustRegister(requestCounter)
}

// IncRequestCounter increments the QPS counter for the given method and path.
func IncRequestCounter(component, path string) {
	requestCounter.WithLabelValues(component, path).Inc()
}
