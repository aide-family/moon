package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_agent/internal/biz/do"
	"prometheus-manager/app/prom_agent/internal/biz/repository"
	"prometheus-manager/pkg/strategy"
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

// SendAlarm send alarm
func (b *AlarmBiz) SendAlarm(alarm *strategy.Alarm) error {
	return b.alarmRepo.Alarm(context.Background(), &do.AlarmDo{
		Alarm: alarm,
	})
}
