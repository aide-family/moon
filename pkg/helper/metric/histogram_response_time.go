package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

var responseTimeHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "api_response_time_seconds",
	Help:    "Histogram for the response time of the API",
	Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 定义时间桶，例如从1ms开始，倍数为2，共15个桶
})

func init() {
	prometheus.MustRegister(responseTimeHistogram)
}

// RecordResponseTime 记录响应时间
func RecordResponseTime(duration float64) {
	responseTimeHistogram.Observe(duration)
}
