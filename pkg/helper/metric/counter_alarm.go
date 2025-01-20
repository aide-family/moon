package metric

import "github.com/prometheus/client_golang/prometheus"

var alarmCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "alarm_total",
		Help:      "count of alarm",
	},
	[]string{"level_id", "strategy_id", "team_id", "strategy_name"},
)

func init() {
	prometheus.MustRegister(alarmCounter)
}

// IncAlarmCounter 告警计数器+1
func IncAlarmCounter(levelID, strategyID, teamID, strategyName string) {
	alarmCounter.WithLabelValues(levelID, strategyID, teamID, strategyName).Inc()
}

// AddAlarmCounter 告警计数器+n
func AddAlarmCounter(levelID, strategyID, teamID, strategyName string, n float64) {
	alarmCounter.WithLabelValues(levelID, strategyID, teamID, strategyName).Add(n)
}
