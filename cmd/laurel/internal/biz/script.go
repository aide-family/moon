package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/bo"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/job"
	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
)

func NewScriptBiz(
	scriptRepo repository.Script,
	eventBusRepo repository.EventBus,
	logger log.Logger,
) *Script {
	return &Script{
		scriptRepo:   scriptRepo,
		eventBusRepo: eventBusRepo,
		helper:       log.NewHelper(log.With(logger, "module", "biz.script")),
	}
}

type Script struct {
	scriptRepo   repository.Script
	eventBusRepo repository.EventBus
	helper       *log.Helper
}

func (s *Script) GetScripts(ctx context.Context) ([]*bo.TaskScript, error) {
	return s.scriptRepo.GetScripts(ctx)
}

func (s *Script) Loads() []cron_server.CronJob {
	return []cron_server.CronJob{
		job.NewLoadFileJob(s.GetScripts, s.eventBusRepo.InScriptJobEventBus, s.eventBusRepo.InRemoveScriptJobEventBus, s.helper.Logger()),
	}
}

func (s *Script) OutScriptJobEventBus() <-chan cron_server.CronJob {
	return s.eventBusRepo.OutScriptJobEventBus()
}

func (s *Script) OutRemoveScriptJobEventBus() <-chan cron_server.CronJob {
	return s.eventBusRepo.OutRemoveScriptJobEventBus()
}
