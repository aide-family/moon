package impl

import (
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/magicbox/server/cron"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/data"
)

func NewJobChannel(d *data.Data) repository.JobChannel {
	metricAppendJobChannel := make(chan cron.CronJob, 100)
	metricRemoveJobChannel := make(chan string, 100)
	jobImpl := &jobChannelRepositoryImpl{
		metricAppendJobChannel: metricAppendJobChannel,
		metricRemoveJobChannel: metricRemoveJobChannel,
	}
	d.AppendClose("jobChannelRepo", jobImpl.close)
	return jobImpl
}

type jobChannelRepositoryImpl struct {
	metricAppendJobChannel chan cron.CronJob
	metricRemoveJobChannel chan string
	closeFuncs             []func() error
}

// AppendMetricJob implements [repository.JobChannel].
func (j *jobChannelRepositoryImpl) AppendMetricJob(job cron.CronJob) {
	j.metricAppendJobChannel <- job
}

func (j *jobChannelRepositoryImpl) close() error {
	close(j.metricAppendJobChannel)
	close(j.metricRemoveJobChannel)
	for _, c := range j.closeFuncs {
		if err := c(); err != nil {
			klog.Errorw("msg", "close job channel failed", "error", err)
		}
	}
	return nil
}

// GetMetricAppendJobChannel implements [repository.JobChannel].
func (j *jobChannelRepositoryImpl) GetMetricAppendJobChannel() <-chan cron.CronJob {
	return j.metricAppendJobChannel
}

// GetMetricRemoveJobChannel implements [repository.JobChannel].
func (j *jobChannelRepositoryImpl) GetMetricRemoveJobChannel() <-chan string {
	return j.metricRemoveJobChannel
}

func (j *jobChannelRepositoryImpl) AppendClose(cs ...func() error) error {
	j.closeFuncs = append(j.closeFuncs, cs...)
	return nil
}
