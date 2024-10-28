package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var counterRequestErr = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "counter_request_err",
		Help: "Counter of request error",
	},
	[]string{"component", "path", "code"},
)

func init() {
	prometheus.MustRegister(counterRequestErr)
}

// IncCounterRequestErr increments the counter of request error.
func IncCounterRequestErr(component, path string, code int32) {
	var status string
	// 5xx error
	if code >= 500 && code < 600 {
		status = "5xx"
	}
	// 4xx error
	if code >= 400 && code < 500 {
		status = "4xx"
	}
	// 3xx error
	if code >= 300 && code < 400 {
		status = "3xx"
	}
	// 2xx error
	if code >= 200 && code < 300 {
		status = "2xx"
	}

	counterRequestErr.WithLabelValues(component, path, status).Inc()
}
