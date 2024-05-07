package dashboard

import (
	"context"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"gorm.io/gorm"
)

var _ repository.ChartRepo = (*chartRepoImpl)(nil)

type chartRepoImpl struct {
	d *data.Data
}

func (l *chartRepoImpl) Create(ctx context.Context, chart *do.MyChart) (*do.MyChart, error) {
	if err := l.db().WithContext(ctx).Create(chart).Error; err != nil {
		return nil, err
	}
	return chart, nil
}

func (l *chartRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*do.MyChart, error) {
	var model do.MyChart
	if err := l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (l *chartRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*do.MyChart, error) {
	var modelList []*do.MyChart
	if err := l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).Find(&modelList).Error; err != nil {
		return nil, err
	}
	return modelList, nil
}

func (l *chartRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*do.MyChart, error) {
	var modelList []*do.MyChart
	wheres := append(scopes, basescopes.WithUserId(ctx))
	if err := l.db().WithContext(ctx).Scopes(append(wheres, bo.Page(pgInfo))...).Find(&modelList).Error; err != nil {
		return nil, err
	}
	var total int64
	if err := l.db().WithContext(ctx).Model(&do.MyChart{}).Scopes(wheres...).Count(&total).Error; err != nil {
		return nil, err
	}
	pgInfo.SetTotal(total)
	return modelList, nil
}

func (l *chartRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	return l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).Delete(&do.MyChart{}).Error
}

func (l *chartRepoImpl) Update(ctx context.Context, chart *do.MyChart, scopes ...basescopes.ScopeMethod) (*do.MyChart, error) {
	wheres := append(scopes, basescopes.WithUserId(ctx), basescopes.InIds(chart.ID))
	if err := l.db().WithContext(ctx).Scopes(wheres...).Updates(chart).Error; err != nil {
		return nil, err
	}
	var first do.MyChart
	if err := l.db().WithContext(ctx).Scopes(wheres...).First(&first).Error; err != nil {
		return nil, err
	}
	return &first, nil
}

func (l *chartRepoImpl) db() *gorm.DB {
	return l.d.DB()
}

func NewChartRepo(d *data.Data) repository.ChartRepo {
	return &chartRepoImpl{
		d: d,
	}
}
