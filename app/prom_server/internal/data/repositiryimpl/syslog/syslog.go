package syslog

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/helper/middler"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.SysLogRepo = (*sysLogRepoImpl)(nil)

type sysLogRepoImpl struct {
	repository.UnimplementedSysLogRepo
	log  *log.Helper
	data *data.Data
}

func NewSysLogRepo(data *data.Data, logger log.Logger) repository.SysLogRepo {
	return &sysLogRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data.syslog")),
	}
}

func (l *sysLogRepoImpl) CreateSysLog(ctx context.Context, action vobj.Action, logInfo ...*bo.SysLogBo) {
	userId := middler.GetUserId(ctx)
	list := slices.To(logInfo, func(detail *bo.SysLogBo) *do.SysLog {
		item := detail.ToModel()
		item.UserId = userId
		item.Action = action
		return item
	})

	l.data.DB().WithContext(ctx).CreateInBatches(list, 100)
}

func (l *sysLogRepoImpl) ListSysLog(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.SysLogBo, error) {
	var logList []*do.SysLog
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, bo.Page(pgInfo))...).Find(&logList).Error; err != nil {
		return nil, err
	}

	if pgInfo != nil {
		var total int64
		if err := l.data.DB().WithContext(ctx).Model(&do.SysLog{}).Scopes(scopes...).Count(&total).Error; err != nil {
			return nil, err
		}
		pgInfo.SetTotal(total)
	}

	list := slices.To(logList, func(detail *do.SysLog) *bo.SysLogBo {
		return bo.SysLogModelToBo(detail)
	})
	return list, nil
}
