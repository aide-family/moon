package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
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
func (b *AlarmBiz) SendAlarm(alarm ...*strategy.Alarm) error {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	for _, item := range alarm {
		alarmInfo := item
		eg.Go(func() error {
			return b.alarmRepo.Alarm(context.Background(), &do.AlarmDo{
				Alarm: alarmInfo,
			})
		})
	}
	return eg.Wait()
}
