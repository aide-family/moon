package controller

import (
	"time"

	"github.com/aide-family/moon/api/cluster/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	namespace        = "cluster"
	moduleController = "controller"
)

// TODO: Complete collect metrics @sumeng

func Metrics() HandlerFunc {
	return func(c *Context) (*time.Duration, error) {
		defer func() {
			phaseGauge(c)
		}()
		// Process request
		return c.Next()
	}
}

func phaseGauge(c *Context) {
	PhaseGauge.DeletePartialMatch(map[string]string{
		"namespace": c.Key.Namespace,
		"name":      c.Key.Name,
	})
	if c.Phase != v1beta1.ClusterPhaseTerminating {
		PhaseGauge.WithLabelValues(c.Key.Name).Set(PhaseCode(c.Status.Phase))
	}
}

var (
	PhaseGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: moduleController,
			Name:      "status_phase_gauge",
			Help:      "cluster status phase (0:Unknown, 1:Initial, 2:Running, 3:Terminating)",
		},
		[]string{"name"},
	)
)

func init() {
	metrics.Registry.MustRegister(PhaseGauge)
}
