package impl

import (
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

const defaultAlertEventChannelCapacity = 1000

func NewAlertEventChannelRepository(bc *conf.Bootstrap, d *data.Data) repository.AlertEventChannel {
	cap := defaultAlertEventChannelCapacity
	if cfg := bc.GetEvaluateConfig(); cfg != nil && cfg.GetAlertEventChannelCapacity() > 0 {
		cap = int(cfg.GetAlertEventChannelCapacity())
	}
	ch := make(chan *bo.AlertEventBo, cap)
	alertEventChannelRepo := &alertEventChannelRepository{ch: ch}
	d.AppendClose("alertEventChannel", alertEventChannelRepo.close)
	return alertEventChannelRepo
}

type alertEventChannelRepository struct {
	ch chan *bo.AlertEventBo
}

func (a *alertEventChannelRepository) Send(event *bo.AlertEventBo) {
	select {
	case a.ch <- event:
	default:
		klog.Warnw("msg", "alert event channel full, dropping event", "strategyUID", event.StrategyUID.Int64())
	}
}

func (a *alertEventChannelRepository) GetChannel() <-chan *bo.AlertEventBo {
	return a.ch
}

func (a *alertEventChannelRepository) close() error {
	close(a.ch)
	return nil
}
