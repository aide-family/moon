package job

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

type GetScriptsFunc func(ctx context.Context) ([]*bo.TaskScript, error)
type InScriptEventBusFunc func(cron_server.CronJob)
type RemoveScriptFunc func(cron_server.CronJob)

func NewLoadFileJob(
	getScripts GetScriptsFunc,
	inScriptEventBus InScriptEventBusFunc,
	removeScript RemoveScriptFunc,
	logger log.Logger,
) cron_server.CronJob {
	return &LoadFileJob{
		getScripts:       getScripts,
		inScriptEventBus: inScriptEventBus,
		removeScript:     removeScript,
		helper:           log.NewHelper(log.With(logger, "module", "job.load_file")),
	}
}

type LoadFileJob struct {
	inScriptEventBus InScriptEventBusFunc
	removeScript     RemoveScriptFunc
	getScripts       GetScriptsFunc
	helper           *log.Helper
	id               cron.EntryID
}

// ID implements cron_server.CronJob.
func (l *LoadFileJob) ID() cron.EntryID {
	return l.id
}

// Index implements cron_server.CronJob.
func (l *LoadFileJob) Index() string {
	return "load_file"
}

// IsImmediate implements cron_server.CronJob.
func (l *LoadFileJob) IsImmediate() bool {
	return true
}

// Run implements cron_server.CronJob.
func (l *LoadFileJob) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	scripts, err := l.getScripts(ctx)
	if err != nil {
		l.helper.Errorf("get scripts error: %v", err)
		return
	}
	l.helper.Infof("load %d scripts", len(scripts))
	for _, script := range scripts {
		job := NewScriptJob(script, l.helper.Logger())
		if script.IsDeleted() {
			l.removeScript(job)
			continue
		}
		l.inScriptEventBus(job)
	}
}

// Spec implements cron_server.CronJob.
func (l *LoadFileJob) Spec() cron_server.CronSpec {
	return cron_server.CronSpecEvery(1 * time.Minute)
}

// WithID implements cron_server.CronJob.
func (l *LoadFileJob) WithID(id cron.EntryID) cron_server.CronJob {
	l.id = id
	return l
}
