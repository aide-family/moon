package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

type (
	// AlarmHistoryBiz 告警历史相关业务逻辑
	AlarmHistoryBiz struct {
		historyRepository repository.HistoryRepository
	}
)

// NewAlarmHistoryBiz 创建告警历史业务逻辑
func NewAlarmHistoryBiz(historyRepository repository.HistoryRepository) *AlarmHistoryBiz {
	return &AlarmHistoryBiz{
		historyRepository: historyRepository,
	}
}

// GetAlarmHistory 获取告警历史
func (a *AlarmHistoryBiz) GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error) {
	return a.historyRepository.GetAlarmHistory(ctx, param)
}

// ListAlarmHistories 获取告警历史列表
func (a *AlarmHistoryBiz) ListAlarmHistories(ctx context.Context, param *bo.QueryAlarmHistoryListParams) ([]*alarmmodel.AlarmHistory, error) {
	return a.historyRepository.GetAlarmHistories(ctx, param)
}
