package metric

import "github.com/prometheus/client_golang/prometheus"

var notifyCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "notify_total",
		Help:      "count of notify",
	},
	[]string{"team_id", "status", "notify_id", "notify_name"},
)

func init() {
	prometheus.MustRegister(notifyCounter)
}

// IncNotifyCounter 通知计数器+1
func IncNotifyCounter(teamID, notifyStatus, notifyID, notifyName string) {
	notifyCounter.WithLabelValues(teamID, notifyStatus, notifyID, notifyName).Inc()
}
