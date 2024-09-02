package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// AlarmHook 告警hook
	AlarmHook interface {
		// CreateAlarmHook 创建告警hook
		CreateAlarmHook(ctx context.Context, params *bo.CreateAlarmHookParams) (*bizmodel.AlarmHook, error)
		// UpdateAlarmHook 更新告警hook
		UpdateAlarmHook(ctx context.Context, params *bo.UpdateAlarmHookParams) error
		// DeleteAlarmHook 删除告警hook
		DeleteAlarmHook(ctx context.Context, ID uint32) error
		// GetAlarmHook 获取告警hook
		GetAlarmHook(ctx context.Context, ID uint32) (*bizmodel.AlarmHook, error)
		// ListAlarmHook 获取告警hook列表
		ListAlarmHook(ctx context.Context, params *bo.QueryAlarmHookListParams) ([]*bizmodel.AlarmHook, error)
		// UpdateAlarmHookStatus 更新告警hook状态
		UpdateAlarmHookStatus(ctx context.Context, params *bo.UpdateAlarmHookStatusParams) error
	}
)
