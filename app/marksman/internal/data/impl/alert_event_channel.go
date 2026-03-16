package impl

import (
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

const defaultAlertEventChannelCapacity = 1000

func NewAlertEventChannel(bc *conf.Bootstrap, d *data.Data) repository.AlertEventChannel {
	cap := defaultAlertEventChannelCapacity
	if cfg := bc.GetEvaluateConfig(); cfg != nil && cfg.GetAlertEventChannelCapacity() > 0 {
		cap = int(cfg.GetAlertEventChannelCapacity())
	}
	ch := make(chan *bo.AlertEventBo, cap)
	impl := &alertEventChannelImpl{ch: ch}
	d.AppendClose("alertEventChannel", impl.close)
	return impl
}

type alertEventChannelImpl struct {
	ch chan *bo.AlertEventBo
}

func (a *alertEventChannelImpl) Send(event *bo.AlertEventBo) {
	select {
	case a.ch <- event:
	default:
		klog.Warnw("msg", "alert event channel full, dropping event", "strategyUID", event.StrategyUID.Int64())
	}
}

func (a *alertEventChannelImpl) GetChannel() <-chan *bo.AlertEventBo {
	return a.ch
}

func (a *alertEventChannelImpl) close() error {
	close(a.ch)
	return nil
}
