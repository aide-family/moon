package bo

import (
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

type StrategyJob interface {
	cron_server.CronJob
	GetEnable() bool
}
