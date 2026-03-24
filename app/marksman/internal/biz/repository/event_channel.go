package repository

import (
	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/biz/bo"
)

type EvaluateJobChannel interface {
	AppendClose(cs ...func() error) error
	AppendEvaluateJob(cron.CronJob)
	RemoveEvaluateJob(index string)
	GetEvaluateJobAppendChannel() <-chan cron.CronJob
	GetEvaluateJobRemoveChannel() <-chan string
}

// AlertEventChannel is a channel that collects alert events from metric evaluators.
type AlertEventChannel interface {
	// Send sends an alert event (non-blocking best-effort; drops if full).
	Send(event *bo.AlertEventBo)
	// GetChannel returns the read-only channel for consumers.
	GetChannel() <-chan *bo.AlertEventBo
}

type AlertingEventChannel interface {
	Append(job cron.CronJob)
	Remove(index string)
	GetJobChannel() <-chan cron.CronJob
	GetRemoveJobChannel() <-chan string
}
