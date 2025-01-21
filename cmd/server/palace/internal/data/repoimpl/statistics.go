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
	latestEventsKey = "palace:events:latest"
)

// getLatestEventsKeyByTeamID 获取最新事件的key
func getLatestEventsKeyByTeamID(teamID string) string {
	return types.TextJoin(latestEventsKey, ":", teamID)
}

// getLatestEventskey 获取最新事件的key
func getLatestEventsKey(teamID uint32) string {
	return types.TextJoin(latestEventsKey, ":", strconv.Itoa(int(teamID)))
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
