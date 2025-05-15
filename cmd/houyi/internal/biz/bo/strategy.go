package bo

import (
	"github.com/aide-family/moon/pkg/plugin/server"
)

type StrategyJob interface {
	server.CronJob
	GetEnable() bool
}
