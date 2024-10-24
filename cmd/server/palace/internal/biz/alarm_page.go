package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// NewAlarmPageBiz 创建告警页面管理功能
func NewAlarmPageBiz(alarmPageRepository repository.AlarmPage) *AlarmPageBiz {
	return &AlarmPageBiz{
		alarmPageRepository: alarmPageRepository,
	}
}

// AlarmPageBiz 告警页面管理功能
type AlarmPageBiz struct {
	alarmPageRepository repository.AlarmPage
}

// UpdateAlarmPage 更新告警页面
func (b *AlarmPageBiz) UpdateAlarmPage(ctx context.Context, userID uint32, alarmPageIDs []uint32) error {
	return b.alarmPageRepository.ReplaceAlarmPages(ctx, userID, alarmPageIDs)
}

// ListAlarmPage 告警页面列表
func (b *AlarmPageBiz) ListAlarmPage(ctx context.Context, userID uint32) ([]*bizmodel.AlarmPageSelf, error) {
	return b.alarmPageRepository.ListAlarmPages(ctx, userID)
}

// GetAlertCounts 获取告警数量
func (b *AlarmPageBiz) GetAlertCounts(ctx context.Context, pageIDs []uint32) map[uint32]int64 {
	return b.alarmPageRepository.GetAlertCounts(ctx, pageIDs)
}
