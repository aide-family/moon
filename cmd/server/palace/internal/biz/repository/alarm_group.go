package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	// AlarmGroup 告警组接口
	AlarmGroup interface {
		// CreateAlarmGroup 创建告警组
		CreateAlarmGroup(context.Context, *bo.CreateAlarmNoticeGroupParams) (*bizmodel.AlarmNoticeGroup, error)
		// UpdateAlarmGroup 更新告警组
		UpdateAlarmGroup(context.Context, *bo.UpdateAlarmNoticeGroupParams) error
		// DeleteAlarmGroup 删除告警组
		DeleteAlarmGroup(context.Context, uint32) error
		// GetAlarmGroup 获取告警详情
		GetAlarmGroup(context.Context, uint32) (*bizmodel.AlarmNoticeGroup, error)
		// AlarmGroupPage 告警列表
		AlarmGroupPage(context.Context, *bo.QueryAlarmNoticeGroupListParams) ([]*bizmodel.AlarmNoticeGroup, error)
		// UpdateStatus 更新状态
		UpdateStatus(context.Context, *bo.UpdateAlarmNoticeGroupStatusParams) error
	}
)
