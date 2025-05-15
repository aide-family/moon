package bo

import (
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

type StrategyJob interface {
	server.CronJob
	GetEnable() bool
}
