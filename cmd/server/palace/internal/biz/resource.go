package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewResourceBiz(resourceRepo repository.Resource) *ResourceBiz {
	return &ResourceBiz{
		resourceRepo: resourceRepo,
	}
}

type ResourceBiz struct {
	resourceRepo repository.Resource
}

// GetResource 获取资源详情
func (b *ResourceBiz) GetResource(ctx context.Context, id uint32) (*model.SysAPI, error) {
	resource, err := b.resourceRepo.GetById(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nResourceNotFoundErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return resource, nil
}

// ListResource 获取资源列表
func (b *ResourceBiz) ListResource(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	resourceDos, err := b.resourceRepo.FindByPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return resourceDos, nil
}

func (b *ResourceBiz) UpdateResourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	if err := b.resourceRepo.UpdateStatus(ctx, status, ids...); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

func (b *ResourceBiz) GetResourceSelectList(ctx context.Context, params *bo.QueryResourceListParams) ([]*bo.SelectOptionBo, error) {
	resourceDos, err := b.resourceRepo.FindSelectByPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}

	return types.SliceTo(resourceDos, func(resource *model.SysAPI) *bo.SelectOptionBo {
		return bo.NewResourceSelectOptionBuild(resource).ToSelectOption()
	}), nil
}
