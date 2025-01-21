package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
)

func NewStatisticsRepository(data *data.Data) repository.Statistics {
	return &statisticsRepositoryImpl{data: data}
}

type statisticsRepositoryImpl struct {
	data *data.Data
}

func (s *statisticsRepositoryImpl) AddEvents(ctx context.Context, events ...*bo.LatestAlarmEvent) error {
	//TODO implement me
	panic("implement me")
}

func (s *statisticsRepositoryImpl) GetLatestEvents(ctx context.Context, teamID uint32, limit int) ([]*bo.LatestAlarmEvent, error) {
	//TODO implement me
	panic("implement me")
}
