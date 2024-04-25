package biz

import (
	"context"

	"github.com/aide-family/moon/app/prom_agent/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_agent/internal/biz/do"
	"github.com/aide-family/moon/app/prom_agent/internal/biz/repository"
	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
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

// SendAlarmV2 send alarm
func (b *AlarmBiz) SendAlarmV2(ctx context.Context, alarm ...*agent.Alarm) error {
	eg := new(errgroup.Group)
	eg.SetLimit(100)
	for _, item := range alarm {
		alarmInfo := item
		eg.Go(func() error {
			return b.alarmRepo.AlarmV2(ctx, &bo.AlarmItemBo{
				Alarm: alarmInfo,
			})
		})
	}
	return eg.Wait()
}
