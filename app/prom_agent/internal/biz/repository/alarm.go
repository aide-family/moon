package repository

import (
	"context"

	"prometheus-manager/app/prom_agent/internal/biz/do"
)

type AlarmRepo interface {
	// Alarm 告警
	Alarm(context.Context, *do.AlarmDo) error
}
