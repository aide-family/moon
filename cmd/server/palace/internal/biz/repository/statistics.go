package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
)

// Statistics 统计数据
type Statistics interface {
	// AddEvents 添加事件
	AddEvents(ctx context.Context, events ...*bo.LatestAlarmEvent) error

	// GetLatestEvents 获取最新事件
	GetLatestEvents(ctx context.Context, teamID uint32, limit int) ([]*bo.LatestAlarmEvent, error)

	// AddInterventionEvents 添加干预事件
	AddInterventionEvents(ctx context.Context, events ...*bo.LatestInterventionEvent) error

	// GetLatestInterventionEvents 获取最新干预事件
	GetLatestInterventionEvents(ctx context.Context, teamID uint32, limit int) ([]*bo.LatestInterventionEvent, error)
}
