package data

// 数据源驱动取值与 ent 的枚举保持一致。手写代码只引用这里的常量，
// 避免在状态检查和查询分发里散落字符串。
const (
	driverMetricsPrometheus      = "METRICS_PROMETHEUS"
	driverMetricsVictoriaMetrics = "METRICS_VICTORIA_METRICS"
	driverLogsLoki               = "LOGS_LOKI"
	driverLogsElasticsearch      = "LOGS_ELASTICSEARCH"
)

// isMetricsDriver 判断驱动是否走 Prometheus 兼容的指标查询接口。
func isMetricsDriver(driver string) bool {
	switch driver {
	case driverMetricsPrometheus, driverMetricsVictoriaMetrics:
		return true
	default:
		return false
	}
}
