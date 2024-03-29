package alarmpage

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.PageRepo = (*alarmPageRepoImpl)(nil)

type alarmPageRepoImpl struct {
	repository.UnimplementedPageRepo
	log *log.Helper

	data *data.Data
}

func (l *alarmPageRepoImpl) UserPageList(ctx context.Context, userId uint32) ([]*bo.DictBO, error) {
	var userInfo do.SysUser
	if err := l.data.DB().
		WithContext(ctx).
		Scopes(do.SysUserPreloadAlarmPages(), basescopes.InIds(userId)).
		First(&userInfo).
		Error; err != nil {
		return nil, err
	}
	return slices.To(userInfo.GetAlarmPages(), func(p *do.SysDict) *bo.DictBO {
		return bo.DictModelToBO(p)
	}), nil
}

func (l *alarmPageRepoImpl) BindUserPages(ctx context.Context, userId uint32, pageIds []uint32) error {
	pagesDo := slices.To(pageIds, func(id uint32) *do.SysDict {
		return &do.SysDict{BaseModel: do.BaseModel{ID: id}}
	})
	return l.data.DB().WithContext(ctx).
		Model(&do.SysUser{BaseModel: do.BaseModel{ID: userId}}).
		Association(do.SysUserPreloadFieldAlarmPages).Replace(pagesDo)
}

func (l *alarmPageRepoImpl) GetPromStrategyAlarmPage(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.PromStrategyAlarmPage, error) {
	var m []*do.PromStrategyAlarmPage
	if err := l.data.DB().
		WithContext(ctx).
		Scopes(scopes...).
		Find(&m).
		Error; err != nil {
		return nil, err
	}
	return m, nil
}

func (l *alarmPageRepoImpl) GetStrategyIds(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]uint32, error) {
	var strategyIds []uint32
	if err := l.data.DB().
		Model(&do.PromStrategyAlarmPage{}).
		WithContext(ctx).
		Scopes(scopes...).
		Pluck(do.PromStrategyAlarmPageFieldPromStrategyID, &strategyIds).
		Error; err != nil {
		return nil, err
	}
	return strategyIds, nil
}

// NewAlarmPageRepo .
func NewAlarmPageRepo(d *data.Data, logger log.Logger) repository.PageRepo {
	return &alarmPageRepoImpl{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data.alarmPage")),
	}
}
