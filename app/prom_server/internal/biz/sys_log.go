package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type SysLogBiz struct {
	log *log.Helper

	repository.SysLogRepo
}

func NewSysLogBiz(repo repository.SysLogRepo, logger log.Logger) *SysLogBiz {
	return &SysLogBiz{
		log:        log.NewHelper(log.With(logger, "module", "biz.sys_log")),
		SysLogRepo: repo,
	}
}

// ListSysLog 获取日志列表
func (b *SysLogBiz) ListSysLog(ctx context.Context, req *bo.ListSyslogReq) ([]*bo.SysLogBo, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.CreatedAtDesc(),
		basescopes.UpdateAtDesc(),
		do.SysLogPreloadUsers(),
		do.SysLogWhereModule(req.Module, req.ModuleId),
	}
	return b.SysLogRepo.ListSysLog(ctx, req.Page, wheres...)
}
