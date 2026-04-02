package cron

import (
	"context"
	"strings"
	"time"

	mcron "github.com/aide-family/magicbox/server/cron"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/conf"
)

func NewCollectSelfJob(bc *conf.Bootstrap, machineInfoBiz *biz.MachineInfo, helper *klog.Helper) *collectSelfJob {
	enabledCollectSelf, collectSelfInterval, collectSelfTimeout := false, 60*time.Second, 60*time.Second
	if collectSelfCfg := bc.GetCollectSelf(); collectSelfCfg != nil {
		enabledCollectSelf = strings.EqualFold(collectSelfCfg.GetEnabled(), "true")
		collectSelfInterval = collectSelfCfg.GetInterval().AsDuration()
		collectSelfTimeout = collectSelfCfg.GetTimeout().AsDuration()
	}

	return &collectSelfJob{
		enabled:     enabledCollectSelf,
		index:       "jade-tree-collect-self",
		spec:        mcron.CronSpecEvery(collectSelfInterval),
		isImmediate: true,
		machineInfo: machineInfoBiz,
		helper:      helper,
		timeout:     collectSelfTimeout,
	}
}

type collectSelfJob struct {
	enabled     bool
	index       string
	spec        mcron.CronSpec
	isImmediate bool
	machineInfo *biz.MachineInfo
	helper      *klog.Helper
	timeout     time.Duration
}

func (j *collectSelfJob) Index() string        { return j.index }
func (j *collectSelfJob) Spec() mcron.CronSpec { return j.spec }
func (j *collectSelfJob) IsImmediate() bool    { return j.isImmediate }
func (j *collectSelfJob) Run() {
	if !j.enabled {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), j.timeout)
	defer cancel()
	j.collectSelfOnce(ctx)
}

func (j *collectSelfJob) collectSelfOnce(ctx context.Context) {
	mi, err := j.machineInfo.RefreshLocalMachineInfo(ctx)
	if err != nil {
		j.helper.Errorw("msg", "collect self failed", "error", err)
		return
	}
	j.helper.Debugw("msg", "collect self success", "machineInfo", mi)
}
