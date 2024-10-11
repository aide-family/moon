package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

type (
	// AlarmRaw repository
	AlarmRaw interface {
		// CreateAlarmRaws 创建告警原始数据
		CreateAlarmRaws(ctx context.Context, params []*bo.CreateAlarmRawParams, teamID uint32) ([]*alarmmodel.AlarmRaw, error)
	}
)
