package repoimpl

import (
	"context"
	"strconv"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/types"
)

func NewStatisticsRepository(data *data.Data) repository.Statistics {
	return &statisticsRepositoryImpl{data: data}
}

type statisticsRepositoryImpl struct {
	data *data.Data
}

const (
	latestEventsKey             = "palace:events:latest"
	latestInterventionEventsKey = "palace:intervention:events:latest"
)

// getLatestEventsKeyByTeamID 获取最新事件的key
func getLatestEventsKeyByTeamID(teamID string) string {
	return types.TextJoin(latestEventsKey, ":", teamID)
}

// getLatestEventskey 获取最新事件的key
func getLatestEventsKey(teamID uint32) string {
	return types.TextJoin(latestEventsKey, ":", strconv.Itoa(int(teamID)))
}

// getLatestInterventionEventsKey 获取最新干预事件的key
func getLatestInterventionEventsKey(teamID uint32) string {
	return types.TextJoin(latestInterventionEventsKey, ":", strconv.Itoa(int(teamID)))
}

// getLatestInterventionEventsKeyByTeamID 获取最新干预事件的key
func getLatestInterventionEventsKeyByTeamID(teamID string) string {
	return types.TextJoin(latestInterventionEventsKey, ":", teamID)
}

// AddEvents 添加事件
func (s *statisticsRepositoryImpl) AddEvents(ctx context.Context, events ...*bo.LatestAlarmEvent) error {
	if len(events) == 0 {
		return nil
	}
	pipe := s.data.GetCacher().Client().Pipeline()
	for _, event := range events {
		pipe.LPush(ctx, getLatestEventsKeyByTeamID(event.TeamID), event)
		pipe.LTrim(ctx, getLatestEventsKeyByTeamID(event.TeamID), 0, 99)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (s *statisticsRepositoryImpl) GetLatestEvents(ctx context.Context, teamID uint32, limit int) ([]*bo.LatestAlarmEvent, error) {
	var events []*bo.LatestAlarmEvent
	err := s.data.GetCacher().Client().
		LRange(ctx, getLatestEventsKey(teamID), 0, int64(limit-1)).
		ScanSlice(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// AddInterventionEvents 添加干预事件
func (s *statisticsRepositoryImpl) AddInterventionEvents(ctx context.Context, events ...*bo.LatestInterventionEvent) error {
	if len(events) == 0 {
		return nil
	}
	pipe := s.data.GetCacher().Client().Pipeline()
	for _, event := range events {
		pipe.LPush(ctx, getLatestInterventionEventsKeyByTeamID(event.TeamID), event)
		pipe.LTrim(ctx, getLatestInterventionEventsKeyByTeamID(event.TeamID), 0, 99)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// GetLatestInterventionEvents 获取最新干预事件
func (s *statisticsRepositoryImpl) GetLatestInterventionEvents(ctx context.Context, teamID uint32, limit int) ([]*bo.LatestInterventionEvent, error) {
	var events []*bo.LatestInterventionEvent
	err := s.data.GetCacher().Client().
		LRange(ctx, getLatestInterventionEventsKey(teamID), 0, int64(limit-1)).
		ScanSlice(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}
