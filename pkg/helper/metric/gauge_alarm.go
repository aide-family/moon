package metric

import "github.com/prometheus/client_golang/prometheus"

var alarmGauge = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "alarm_realtime_total",
		Help:      "count of alarm",
	},
	[]string{"level_id", "strategy_id", "team_id"},
)

func init() {
	prometheus.MustRegister(alarmGauge)
}

// IncAlarmGauge 告警计数器+1
func IncAlarmGauge(levelID, strategyID, teamID string) {
	alarmGauge.WithLabelValues(levelID, strategyID, teamID).Inc()
}

// DecAlarmGauge 告警计数器-1
func DecAlarmGauge(levelID, strategyID, teamID string) {
	alarmGauge.WithLabelValues(levelID, strategyID, teamID).Dec()
}
