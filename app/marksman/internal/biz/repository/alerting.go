package repository

import (
	"github.com/aide-family/magicbox/server/cron"
)

type Alerting interface {
	Append(job cron.CronJob)
	Remove(index string)
	GetJobChannel() <-chan cron.CronJob
	GetRemoveJobChannel() <-chan string
}
