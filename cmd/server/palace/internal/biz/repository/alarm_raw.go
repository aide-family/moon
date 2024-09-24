package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

type (
	// AlarmRaw repository
	AlarmRaw interface {
		// CreateAlarmRaw 创建报警规则
		CreateAlarmRaw(ctx context.Context, params *bo.CreateAlarmRawParams) (*alarmmodel.AlarmRaw, error)
	}
)
