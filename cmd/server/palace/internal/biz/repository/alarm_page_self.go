package repository

import (
	"context"

	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// AlarmPage 告警页面管理
type AlarmPage interface {
	// ReplaceAlarmPages 批量替换告警页面
	ReplaceAlarmPages(ctx context.Context, userID uint32, alarmPageIDs []uint32) error

	// ListAlarmPages 获取用户告警页面列表
	ListAlarmPages(ctx context.Context, userID uint32) ([]*bizmodel.AlarmPageSelf, error)

	// GetAlertCounts 获取告警页面的告警数量
	GetAlertCounts(ctx context.Context, pageIDs []uint32) map[int32]int64
}
