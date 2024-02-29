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

	// UPMemberCounter 用户在线人数
	UPMemberCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "member_total",
		Help:      "The total number of processed requests",
	}, []string{"server"})

	// AlarmEventCounter 告警事件数
	AlarmEventCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "alarm_total",
		Help:      "The total number of processed requests",
	}, []string{"strategy_id"})

	// WorkingStrategyCounter 工作中规则数量
	WorkingStrategyCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "server",
		Subsystem: "requests",
		Name:      "working_strategy_total",
		Help:      "The total number of processed requests",
	}, []string{"server"})
)

func init() {
	prometheus.MustRegister(
		MetricSeconds,
		MetricRequests,
		IpMetricCounter,
		UPMemberCounter,
		AlarmEventCounter,
		WorkingStrategyCounter,
	)
}
