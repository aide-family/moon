package biz

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
)

func NewAlarmLevelBiz(alarmLevelRepository repository.AlarmLevel) *AlarmLevelBiz {
	return &AlarmLevelBiz{
		alarmLevelRepository: alarmLevelRepository,
	}
}

// AlarmLevelBiz .
type AlarmLevelBiz struct {
	alarmLevelRepository repository.AlarmLevel
}
