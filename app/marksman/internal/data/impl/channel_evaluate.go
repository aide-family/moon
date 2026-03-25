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
//
// TODO:
//
//   - Issue: AppendEvaluateJob 在 select 的 default 分支仅打日志，不向调用方返回错误，也不阻塞重试。高负载或启动瞬间大量 localStrategyMetricsByNamespace 投递时，定时评估任务会被静默丢弃，存在漏告警风险。
//
//   - Recommendation: 至少任选其一并贯彻：增大 appendChannelCapacity 默认值并监控 channel 深度；default 改为带退避的重试或阻塞发送（需配合 shutdown 避免死锁）；或将 AppendEvaluateJob 改为 error 返回，由上游记录/指标告警。避免仅 Warn 且无补偿。
func (j *evaluateJobChannelRepository) AppendEvaluateJob(job cron.CronJob) {
	select {
	case j.metricAppendJobChannel <- job:
	default:
		klog.Warnw("msg", "evaluate job channel full, dropping job", "job", job.Index())
	}
}

func (j *evaluateJobChannelRepository) RemoveEvaluateJob(index string) {
	select {
	case j.metricRemoveJobChannel <- index:
	default:
		klog.Warnw("msg", "evaluate remove job channel full, dropping job", "index", index)
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
