package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
)

func NewStatisticsBiz(statisticsRepository repository.Statistics) *StatisticsBiz {
	return &StatisticsBiz{
		statisticsRepository: statisticsRepository,
	}
}

// StatisticsBiz .
type StatisticsBiz struct {
	statisticsRepository repository.Statistics
}

// AddEvents 添加事件
func (s *StatisticsBiz) AddEvents(ctx context.Context, events ...*bo.LatestAlarmEvent) error {
	return s.statisticsRepository.AddEvents(ctx, events...)
}

// GetLatestEvents 获取最新事件
func (s *StatisticsBiz) GetLatestEvents(ctx context.Context, limit int) ([]*bo.LatestAlarmEvent, error) {
	return s.statisticsRepository.GetLatestEvents(ctx, middleware.GetTeamID(ctx), limit)
}
