package repository

import (
	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
)

type EventBus interface {
	InStrategyJobEventBus() chan<- bo.StrategyJob
	OutStrategyJobEventBus() <-chan bo.StrategyJob

	InAlertJobEventBus() chan<- bo.AlertJob
	OutAlertJobEventBus() <-chan bo.AlertJob

	InAlertEventBus() chan<- bo.Alert
	OutAlertEventBus() <-chan bo.Alert
}
