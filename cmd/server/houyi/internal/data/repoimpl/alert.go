package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
)

// NewAlertRepository 实例化alert
func NewAlertRepository(data *data.Data) repository.Alert {
	return &alertRepositoryImpl{data: data}
}

type alertRepositoryImpl struct {
	data *data.Data
}

func (a *alertRepositoryImpl) SaveAlarm(_ context.Context, alarm *bo.Alarm) error {
	return a.data.GetAlertQueue().Push(alarm.Message())
}
