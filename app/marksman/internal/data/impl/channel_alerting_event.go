package impl

import (
	"github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
)

const (
	defaultAlertingJobChannelCapacity       = 1000
	defaultAlertingRemoveJobChannelCapacity = 1000
)

func NewAlertingEventChannelRepository(d *data.Data) repository.AlertingEventChannel {
	jobChannel := make(chan cron.CronJob, defaultAlertingJobChannelCapacity)
	removeJobChannel := make(chan string, defaultAlertingRemoveJobChannelCapacity)
	alertingEventChannelRepo := &alertingEventChannelRepository{
		Data:             d,
		jobChannel:       jobChannel,
		removeJobChannel: removeJobChannel,
	}
	d.AppendClose("alertingEventChannel", alertingEventChannelRepo.close)
	return alertingEventChannelRepo
}

type alertingEventChannelRepository struct {
	*data.Data
	jobChannel       chan cron.CronJob
	removeJobChannel chan string
}

// Append implements [repository.Alerting].
func (a *alertingEventChannelRepository) Append(job cron.CronJob) {
	select {
	case a.jobChannel <- job:
	default:
		klog.Warnw("msg", "alerting job channel full, dropping job", "job", job.Index())
	}
}

func (a *alertingEventChannelRepository) Remove(index string) {
	select {
	case a.removeJobChannel <- index:
	default:
		klog.Warnw("msg", "alerting remove job channel full, dropping job", "index", index)
	}
}

func (a *alertingEventChannelRepository) GetJobChannel() <-chan cron.CronJob {
	return a.jobChannel
}

func (a *alertingEventChannelRepository) GetRemoveJobChannel() <-chan string {
	return a.removeJobChannel
}

func (a *alertingEventChannelRepository) close() error {
	close(a.jobChannel)
	close(a.removeJobChannel)
	return nil
}
