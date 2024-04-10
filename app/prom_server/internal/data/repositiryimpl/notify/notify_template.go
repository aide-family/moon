package notify

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ repository.NotifyTemplateRepo = (*notifyTemplateRepoImpl)(nil)

type notifyTemplateRepoImpl struct {
	repository.UnimplementedNotifyTemplateRepo
	log *log.Helper

	d *data.Data
}

func (n *notifyTemplateRepoImpl) Get(ctx context.Context, scopes ...basescopes.ScopeMethod) (*bo.NotifyTemplateBO, error) {
	var notifyTemplate do.PromStrategyNotifyTemplate
	if err := n.d.DB().WithContext(ctx).Scopes(scopes...).First(&notifyTemplate).Error; err != nil {
		return nil, err
	}
	return bo.NotifyTemplateModelToBO(&notifyTemplate), nil
}

func (n *notifyTemplateRepoImpl) Find(ctx context.Context, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error) {
	var notifyTemplates []*do.PromStrategyNotifyTemplate
	if err := n.d.DB().WithContext(ctx).Scopes(scopes...).Find(&notifyTemplates).Error; err != nil {
		return nil, err
	}
	return slices.To(notifyTemplates, func(t *do.PromStrategyNotifyTemplate) *bo.NotifyTemplateBO {
		return bo.NotifyTemplateModelToBO(t)
	}), nil
}

func (n *notifyTemplateRepoImpl) Count(ctx context.Context, scopes ...basescopes.ScopeMethod) (int64, error) {
	var total int64
	if err := n.d.DB().WithContext(ctx).Model(&do.PromStrategyNotifyTemplate{}).Scopes(scopes...).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (n *notifyTemplateRepoImpl) List(ctx context.Context, pgInfo bo.Pagination, scopes ...basescopes.ScopeMethod) ([]*bo.NotifyTemplateBO, error) {
	var notifyTemplates []*do.PromStrategyNotifyTemplate
	if err := n.d.DB().WithContext(ctx).Scopes(append(scopes, bo.Page(pgInfo))...).Find(&notifyTemplates).Error; err != nil {
		return nil, err
	}
	if pgInfo != nil {
		count, err := n.Count(ctx, scopes...)
		if err != nil {
			return nil, err
		}
		pgInfo.SetTotal(count)
	}
	return slices.To(notifyTemplates, func(t *do.PromStrategyNotifyTemplate) *bo.NotifyTemplateBO {
		return bo.NotifyTemplateModelToBO(t)
	}), nil
}

func (n *notifyTemplateRepoImpl) Create(ctx context.Context, notifyTemplate *bo.NotifyTemplateBO) (*bo.NotifyTemplateBO, error) {
	newModel := notifyTemplate.ToModel()
	if err := n.d.DB().WithContext(ctx).Create(newModel).Error; err != nil {
		return nil, err
	}
	return bo.NotifyTemplateModelToBO(newModel), nil
}

func (n *notifyTemplateRepoImpl) Update(ctx context.Context, notifyTemplate *bo.NotifyTemplateBO, scopes ...basescopes.ScopeMethod) error {
	return n.d.DB().WithContext(ctx).Scopes(scopes...).Updates(notifyTemplate.ToModel()).Error
}

func (n *notifyTemplateRepoImpl) Delete(ctx context.Context, scopes ...basescopes.ScopeMethod) error {
	return n.d.DB().WithContext(ctx).Scopes(scopes...).Delete(&do.PromStrategyNotifyTemplate{}).Error
}

func NewNotifyTemplateRepo(d *data.Data, logger log.Logger) repository.NotifyTemplateRepo {
	return &notifyTemplateRepoImpl{
		log: log.NewHelper(logger),
		d:   d,
	}
}
