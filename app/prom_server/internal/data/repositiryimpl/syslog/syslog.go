package syslog

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"
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

func (l *sysLogRepoImpl) CreateSysLog(ctx context.Context, action vo.Action, logInfo ...*bo.SysLogBo) {
	userId := middler.GetUserId(ctx)
	list := slices.To(logInfo, func(detail *bo.SysLogBo) *do.SysLog {
		item := detail.ToModel()
		item.UserId = userId
		item.Action = action
		return item
	})

	l.data.DB().WithContext(ctx).CreateInBatches(list, 100)
}

func (l *sysLogRepoImpl) ListSysLog(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.SysLogBo, error) {
	var logList []*do.SysLog
	if err := l.data.DB().WithContext(ctx).Scopes(append(scopes, basescopes.Page(pgInfo))...).Find(&logList).Error; err != nil {
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
