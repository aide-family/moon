package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
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
func (b *SysLogBiz) ListSysLog(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.SysLogBo, error) {
	return b.SysLogRepo.ListSysLog(ctx, pgInfo, scopes...)
}
