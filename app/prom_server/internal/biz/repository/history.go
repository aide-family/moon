package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	// HistoryRepo .
	HistoryRepo interface {
		// GetHistoryById 通过id获取历史详情
		GetHistoryById(ctx context.Context, id uint) (*dobo.AlarmHistoryDO, error)
		// ListHistory 获取历史列表
		ListHistory(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.AlarmHistoryDO, error)
		// CreateHistory 创建历史
		CreateHistory(ctx context.Context, historyDo *dobo.AlarmHistoryDO) (*dobo.AlarmHistoryDO, error)
		// UpdateHistoryById 通过id更新历史
		UpdateHistoryById(ctx context.Context, id uint, historyDo *dobo.AlarmHistoryDO) (*dobo.AlarmHistoryDO, error)
	}
)
