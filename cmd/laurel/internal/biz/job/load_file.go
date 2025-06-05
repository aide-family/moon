package job

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

type GetScriptsFunc func(ctx context.Context) ([]*bo.TaskScript, error)

func NewLoadFileJob(
	getScripts GetScriptsFunc,
	eventBus repository.EventBus,
	logger log.Logger,
) cron_server.CronJob {
	return &LoadFileJob{
		getScripts: getScripts,
		eventBus:   eventBus,
		helper:     log.NewHelper(log.With(logger, "module", "job.load_file")),
	}
}

type LoadFileJob struct {
	getScripts GetScriptsFunc
	eventBus   repository.EventBus
	helper     *log.Helper
	id         cron.EntryID
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
		job := NewScriptJob(script, l.eventBus, l.helper.Logger())
		if script.IsDeleted() {
			l.eventBus.InRemoveScriptJobEventBus(job)
			continue
		}
		l.eventBus.InScriptJobEventBus(job)
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
