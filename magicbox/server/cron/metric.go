package cron

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cronJobRunCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cronjob_run_count",
			Help: "Total number of cron job runs, by job name and result.",
		},
		[]string{"job_name", "result"}, // success/failure
	)

	cronJobRunDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cronjob_run_duration_seconds",
			Help:    "Histogram of cron job run durations, by job name.",
			Buckets: prometheus.ExponentialBuckets(0.1, 2, 8), // 0.1s ~ 12.8s
		},
		[]string{"job_name"},
	)

	cronJobLastRunTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cronjob_last_run_timestamp",
			Help: "Unix timestamp of the last run of the cron job, by job name.",
		},
		[]string{"job_name"},
	)
)

type instrumentedCronJob struct {
	job CronJob
}

func (i *instrumentedCronJob) Run() {
	jobName := i.job.Index()
	start := time.Now()
	var result string

	defer func() {
		duration := time.Since(start).Seconds()

		cronJobRunDuration.WithLabelValues(jobName).Observe(duration)
		cronJobRunCounter.WithLabelValues(jobName, result).Inc()
		cronJobLastRunTimestamp.WithLabelValues(jobName).Set(float64(time.Now().Unix()))
	}()

	defer func() {
		if r := recover(); r != nil {
			result = "failure"
		}
	}()

	i.job.Run()
	result = "success"
}

func (i *instrumentedCronJob) Index() string     { return i.job.Index() }
func (i *instrumentedCronJob) Spec() CronSpec    { return i.job.Spec() }
func (i *instrumentedCronJob) IsImmediate() bool { return i.job.IsImmediate() }

// WrapCronJobWithMetrics 包装一个 CronJob，使其执行时自动上报 metrics
func WrapCronJobWithMetrics(job CronJob) CronJob {
	return &instrumentedCronJob{job: job}
}
