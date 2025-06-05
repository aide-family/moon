package service

import (
	"github.com/aide-family/moon/cmd/laurel/internal/biz"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/go-kratos/kratos/v2/log"
)

func NewScriptService(scriptBiz *biz.Script, logger log.Logger) *ScriptService {
	return &ScriptService{
		scriptBiz: scriptBiz,
		helper:    log.NewHelper(log.With(logger, "module", "service.script")),
	}
}

type ScriptService struct {
	scriptBiz *biz.Script
	helper    *log.Helper
}

func (s *ScriptService) OutScriptJobEventBus() <-chan cron_server.CronJob {
	return s.scriptBiz.OutScriptJobEventBus()
}

func (s *ScriptService) OutRemoveScriptJobEventBus() <-chan cron_server.CronJob {
	return s.scriptBiz.OutRemoveScriptJobEventBus()
}

func (s *ScriptService) OutMetricEventBus() <-chan []byte {
	return s.scriptBiz.OutMetricEventBus()
}

func (s *ScriptService) Loads() []cron_server.CronJob {
	return s.scriptBiz.Loads()
}
