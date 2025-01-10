package metric

import "github.com/prometheus/client_golang/prometheus"

var alarmCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "alarm_total",
		Help:      "count of alarm",
	},
	[]string{"level_id", "strategy_id", "team_id"},
)

func init() {
	prometheus.MustRegister(alarmCounter)
}

// IncAlarmCounter 告警计数器+1
func IncAlarmCounter(levelID, strategyID, teamID string) {
	alarmCounter.WithLabelValues(levelID, strategyID, teamID).Inc()
}

// AddAlarmCounter 告警计数器+n
func AddAlarmCounter(levelID, strategyID, teamID string, n float64) {
	alarmCounter.WithLabelValues(levelID, strategyID, teamID).Add(n)
}
