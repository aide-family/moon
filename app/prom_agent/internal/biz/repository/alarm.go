package repository

import (
	"context"

	"github.com/aide-family/moon/app/prom_agent/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_agent/internal/biz/do"
)

type AlarmRepo interface {
	// Alarm 告警
	Alarm(context.Context, *do.AlarmDo) error

	// AlarmV2 告警
	AlarmV2(context.Context, *bo.AlarmItemBo) error
}
