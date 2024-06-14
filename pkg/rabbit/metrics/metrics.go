package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	ReceiverTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rabbit_runtime_receive_total",
		Help: "Total number of receiver per rabbit",
	}, []string{"receiver"})

	WorkerTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rabbit_runtime_worker_total",
		Help: "Total number of worker per rabbit",
	}, []string{"result"})

	// WorkerErrors is a prometheus counter metrics which holds the total
	// number of errors from the Worker.
	WorkerErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "rabbit_runtime_worker_errors_total",
		Help: "Total number of worker errors per rabbit",
	}, []string{""})
	// WorkerTime is a prometheus metric which keeps track of the duration
	// of worker.
	WorkerTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "rabbit_runtime_worker_time_seconds",
		Help: "Length of time per worker per rabbit",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
			1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
	}, []string{""})

	// WorkerCount is a prometheus metric which holds the number of
	// concurrent workers per rabbit.
	WorkerCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rabbit_runtime_max_concurrent_workers",
		Help: "Maximum number of concurrent workers per rabbit",
	}, []string{""})

	// ActiveWorkers is a prometheus metric which holds the number
	// of active workers per rabbit.
	ActiveWorkers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rabbit_runtime_active_workers",
		Help: "Number of currently used workers per rabbit",
	}, []string{""})
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(
		ReceiverTotal,
		WorkerTotal,
		WorkerErrors,
		WorkerTime,
		WorkerCount,
		ActiveWorkers,
		// expose process metrics like CPU, Memory, file descriptor usage etc.
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		// expose Go runtime metrics like GC stats, memory stats etc.
		collectors.NewGoCollector(),
	)
}
