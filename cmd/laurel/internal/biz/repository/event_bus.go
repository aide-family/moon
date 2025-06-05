package repository

import (
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

type EventBus interface {
	OutScriptJobEventBus() <-chan cron_server.CronJob
	InScriptJobEventBus(job cron_server.CronJob)
	OutRemoveScriptJobEventBus() <-chan cron_server.CronJob
	InRemoveScriptJobEventBus(job cron_server.CronJob)
	OutMetricEventBus() <-chan []byte
	InMetricEventBus(event []byte)
}
