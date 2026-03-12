package repository

import "github.com/aide-family/magicbox/server/cron"

type JobChannel interface {
	AppendClose(cs ...func() error) error
	AppendMetricJob(cron.CronJob)
	GetMetricAppendJobChannel() <-chan cron.CronJob
	GetMetricRemoveJobChannel() <-chan string
}
