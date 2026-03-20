package impl

import (
	"github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
)

const (
	defaultAppendChannelCapacity = 100
	defaultRemoveChannelCapacity = 100
)

func NewEvaluateJobChannelRepository(bc *conf.Bootstrap, d *data.Data) repository.EvaluateJobChannel {
	appendCap := defaultAppendChannelCapacity
	removeCap := defaultRemoveChannelCapacity
	if cfg := bc.GetEvaluateConfig(); cfg != nil {
		if cfg.GetAppendChannelCapacity() > 0 {
			appendCap = int(cfg.GetAppendChannelCapacity())
		}
		if cfg.GetRemoveChannelCapacity() > 0 {
			removeCap = int(cfg.GetRemoveChannelCapacity())
		}
	}
	metricAppendJobChannel := make(chan cron.CronJob, appendCap)
	metricRemoveJobChannel := make(chan string, removeCap)
	evaluateJobChannelRepo := &evaluateJobChannelRepository{
		metricAppendJobChannel: metricAppendJobChannel,
		metricRemoveJobChannel: metricRemoveJobChannel,
	}
	d.AppendClose("evaluateJobChannelRepo", evaluateJobChannelRepo.close)
	return evaluateJobChannelRepo
}

type evaluateJobChannelRepository struct {
	metricAppendJobChannel chan cron.CronJob
	metricRemoveJobChannel chan string
	closeFuncs             []func() error
}

// AppendEvaluateJob implements [repository.EvaluateJobChannel].
func (j *evaluateJobChannelRepository) AppendEvaluateJob(job cron.CronJob) {
	select {
	case j.metricAppendJobChannel <- job:
	default:
		klog.Warnw("msg", "evaluate job channel full, dropping job", "job", job.Index())
	}
}

func (j *evaluateJobChannelRepository) close() error {
	close(j.metricAppendJobChannel)
	close(j.metricRemoveJobChannel)
	for _, c := range j.closeFuncs {
		if err := c(); err != nil {
			klog.Errorw("msg", "close job channel failed", "error", err)
		}
	}
	return nil
}

// GetEvaluateJobAppendChannel implements [repository.EvaluateJobChannel].
func (j *evaluateJobChannelRepository) GetEvaluateJobAppendChannel() <-chan cron.CronJob {
	return j.metricAppendJobChannel
}

// GetEvaluateJobRemoveChannel implements [repository.EvaluateJobChannel].
func (j *evaluateJobChannelRepository) GetEvaluateJobRemoveChannel() <-chan string {
	return j.metricRemoveJobChannel
}

func (j *evaluateJobChannelRepository) AppendClose(cs ...func() error) error {
	j.closeFuncs = append(j.closeFuncs, cs...)
	return nil
}
