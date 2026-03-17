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

func NewAlertingRepository(d *data.Data) repository.Alerting {
	jobChannel := make(chan cron.CronJob, defaultAlertingJobChannelCapacity)
	removeJobChannel := make(chan string, defaultAlertingRemoveJobChannelCapacity)
	alerting := &alertingRepository{
		Data:             d,
		jobChannel:       jobChannel,
		removeJobChannel: removeJobChannel,
	}
	d.AppendClose("alerting", alerting.close)
	return alerting
}

type alertingRepository struct {
	*data.Data
	jobChannel       chan cron.CronJob
	removeJobChannel chan string
}

// Append implements [repository.Alerting].
func (a *alertingRepository) Append(job cron.CronJob) {
	select {
	case a.jobChannel <- job:
	default:
		klog.Warnw("msg", "alerting job channel full, dropping job", "job", job.Index())
	}
}

func (a *alertingRepository) Remove(index string) {
	select {
	case a.removeJobChannel <- index:
	default:
		klog.Warnw("msg", "alerting remove job channel full, dropping job", "index", index)
	}
}

func (a *alertingRepository) GetJobChannel() <-chan cron.CronJob {
	return a.jobChannel
}

func (a *alertingRepository) GetRemoveJobChannel() <-chan string {
	return a.removeJobChannel
}

func (a *alertingRepository) close() error {
	close(a.jobChannel)
	close(a.removeJobChannel)
	return nil
}
