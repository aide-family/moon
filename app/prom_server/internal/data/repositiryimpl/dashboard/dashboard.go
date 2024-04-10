package dashboard

import (
	"context"

	"gorm.io/gorm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.DashboardRepo = (*dashboardRepoImpl)(nil)

type dashboardRepoImpl struct {
	repository.UnimplementedDashboardRepo
	d *data.Data
}

func (l *dashboardRepoImpl) Create(ctx context.Context, dashboard *bo.MyDashboardConfigBO) (*bo.MyDashboardConfigBO, error) {
	newModel := dashboard.ToModel()
	err := l.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newModel).Error; err != nil {
			return err
		}
		if err := tx.Model(newModel).Association(do.MyDashboardConfigPreloadFieldCharts).Replace(newModel.GetCharts()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return bo.MyDashboardConfigModelToBO(newModel), err
}

func (l *dashboardRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error) {
	var model do.MyDashboardConfig
	wheres := append(scopes, basescopes.WithUserId(ctx))
	err := l.db().WithContext(ctx).Scopes(wheres...).First(&model).Error
	if err != nil {
		return nil, err
	}
	return bo.MyDashboardConfigModelToBO(&model), err
}

func (l *dashboardRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error) {
	var modelList []*do.MyDashboardConfig
	wheres := append(scopes, basescopes.WithUserId(ctx))
	err := l.db().WithContext(ctx).Scopes(wheres...).Find(&modelList).Error
	if err != nil {
		return nil, err
	}
	return slices.To(modelList, func(i *do.MyDashboardConfig) *bo.MyDashboardConfigBO { return bo.MyDashboardConfigModelToBO(i) }), err
}

func (l *dashboardRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.MyDashboardConfigBO, error) {
	var modelList []*do.MyDashboardConfig
	wheres := append(scopes, basescopes.WithUserId(ctx))
	err := l.db().WithContext(ctx).Scopes(append(wheres, bo.Page(pgInfo))...).Find(&modelList).Error
	if err != nil {
		return nil, err
	}
	var total int64
	err = l.d.DB().WithContext(ctx).Model(&do.MyDashboardConfig{}).Scopes(wheres...).Count(&total).Error
	if err != nil {
		return nil, err
	}
	pgInfo.SetTotal(total)
	return slices.To(modelList, func(i *do.MyDashboardConfig) *bo.MyDashboardConfigBO { return bo.MyDashboardConfigModelToBO(i) }), nil
}

func (l *dashboardRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	wheres := append(scopes, basescopes.WithUserId(ctx))
	return l.d.DB().WithContext(ctx).Scopes(wheres...).Delete(&do.MyDashboardConfig{}).Error
}

func (l *dashboardRepoImpl) Update(ctx context.Context, dashboard *bo.MyDashboardConfigBO, scopes ...basescopes.ScopeMethod) (*bo.MyDashboardConfigBO, error) {
	newModel := dashboard.ToModel()
	wheres := append(scopes, basescopes.WithUserId(ctx))
	err := l.d.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(newModel).Scopes(wheres...).Updates(newModel).Error; err != nil {
			return err
		}
		if err := tx.Model(newModel).Association(do.MyDashboardConfigPreloadFieldCharts).Replace(newModel.GetCharts()); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var first do.MyDashboardConfig
	if err = l.d.DB().WithContext(ctx).Scopes(wheres...).First(&first).Error; err != nil {
		return nil, err
	}
	return bo.MyDashboardConfigModelToBO(&first), err
}

func (l *dashboardRepoImpl) db() *gorm.DB {
	return l.d.DB()
}

func NewDashboardRepo(d *data.Data) repository.DashboardRepo {
	return &dashboardRepoImpl{
		d: d,
	}
}
