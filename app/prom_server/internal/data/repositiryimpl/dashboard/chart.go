package dashboard

import (
	"context"

	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.ChartRepo = (*chartRepoImpl)(nil)

type chartRepoImpl struct {
	d *data.Data
	repository.UnimplementedChartRepo
}

func (l *chartRepoImpl) Create(ctx context.Context, chart *bo.MyChartBO) (*bo.MyChartBO, error) {
	newModel := chart.ToModel()
	if err := l.db().WithContext(ctx).Create(newModel).Error; err != nil {
		return nil, err
	}
	return bo.MyChartModelToBO(newModel), nil
}

func (l *chartRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.MyChartBO, error) {
	var model do.MyChart
	if err := l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).First(&model).Error; err != nil {
		return nil, err
	}
	return bo.MyChartModelToBO(&model), nil
}

func (l *chartRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error) {
	var modelList []*do.MyChart
	if err := l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).Find(&modelList).Error; err != nil {
		return nil, err
	}
	return slices.To(modelList, func(i *do.MyChart) *bo.MyChartBO { return bo.MyChartModelToBO(i) }), nil
}

func (l *chartRepoImpl) List(ctx context.Context, pgInfo basescopes.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.MyChartBO, error) {
	var modelList []*do.MyChart
	wheres := append(scopes, basescopes.WithUserId(ctx))
	if err := l.db().WithContext(ctx).Scopes(append(wheres, basescopes.Page(pgInfo))...).Find(&modelList).Error; err != nil {
		return nil, err
	}
	var total int64
	if err := l.db().WithContext(ctx).Model(&do.MyChart{}).Scopes(wheres...).Count(&total).Error; err != nil {
		return nil, err
	}
	pgInfo.SetTotal(total)
	return slices.To(modelList, func(i *do.MyChart) *bo.MyChartBO { return bo.MyChartModelToBO(i) }), nil
}

func (l *chartRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	return l.db().WithContext(ctx).Scopes(append(scopes, basescopes.WithUserId(ctx))...).Delete(&do.MyChart{}).Error
}

func (l *chartRepoImpl) Update(ctx context.Context, chart *bo.MyChartBO, scopes ...basescopes.ScopeMethod) (*bo.MyChartBO, error) {
	newModel := chart.ToModel()
	wheres := append(scopes, basescopes.WithUserId(ctx))
	if err := l.db().WithContext(ctx).Scopes(wheres...).Updates(newModel).Error; err != nil {
		return nil, err
	}
	var first do.MyChart
	if err := l.db().WithContext(ctx).Scopes(wheres...).First(&first).Error; err != nil {
		return nil, err
	}
	return bo.MyChartModelToBO(&first), nil
}

func (l *chartRepoImpl) db() *gorm.DB {
	return l.d.DB()
}

func NewChartRepo(d *data.Data) repository.ChartRepo {
	return &chartRepoImpl{
		d: d,
	}
}
