package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

type (
	// HistoryRepository 告警历史仓库接口
	HistoryRepository interface {
		// GetAlarmHistory 获取告警历史
		GetAlarmHistory(ctx context.Context, param *bo.GetAlarmHistoryParams) (*alarmmodel.AlarmHistory, error)
		// GetAlarmHistories 获取告警历史列表
		GetAlarmHistories(ctx context.Context, param *bo.QueryAlarmHistoryListParams) ([]*alarmmodel.AlarmHistory, error)
		// CreateAlarmHistory 创建告警历史
		CreateAlarmHistory(ctx context.Context, param *bo.CreateAlarmInfoParams) error
	}
)
