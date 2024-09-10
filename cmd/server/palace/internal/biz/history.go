package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

// AlarmHistoryBiz 告警历史相关业务逻辑
type (
	AlarmHistoryBiz struct {
		historyRepository repository.HistoryRepository
	}
)

func NewAlarmHistoryBiz(historyRepository repository.HistoryRepository) *AlarmHistoryBiz {
	return &AlarmHistoryBiz{
		historyRepository: historyRepository,
	}
}

func (a *AlarmHistoryBiz) GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error) {
	return a.historyRepository.GetAlarmHistory(ctx, param)
}

func (a *AlarmHistoryBiz) ListAlarmHistories(ctx context.Context, param *bo.QueryAlarmHistoryListParams) ([]*alarmmodel.AlarmHistory, error) {
	return a.historyRepository.GetAlarmHistories(ctx, param)
}
