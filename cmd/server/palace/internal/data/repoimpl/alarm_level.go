package repoimpl

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
)

func NewAlarmLevelRepository(data *data.Data) repository.AlarmLevel {
	return &alarmLevelRepositoryImpl{data: data}
}

type alarmLevelRepositoryImpl struct {
	data *data.Data
}
