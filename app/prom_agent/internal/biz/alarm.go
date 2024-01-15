package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_agent/internal/biz/repository"
)

type AlarmBiz struct {
	log       *log.Helper
	alarmRepo repository.AlarmRepo
}

func NewAlarmBiz(alarmRepo repository.AlarmRepo, logger log.Logger) *AlarmBiz {
	return &AlarmBiz{
		log:       log.NewHelper(log.With(logger, "module", "biz.alarm")),
		alarmRepo: alarmRepo,
	}
}
