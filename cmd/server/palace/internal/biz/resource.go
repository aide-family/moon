package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewResourceBiz 创建资源业务
func NewResourceBiz(resourceRepo repository.Resource, teamResourceRepo repository.TeamResource) *ResourceBiz {
	return &ResourceBiz{
		resourceRepo:     resourceRepo,
		teamResourceRepo: teamResourceRepo,
	}
}

// ResourceBiz 资源业务
type ResourceBiz struct {
	resourceRepo     repository.Resource
	teamResourceRepo repository.TeamResource
}

func (b *ResourceBiz) getResourceRepo(ctx context.Context) repository.Resource {
	if middleware.GetSourceType(ctx).IsSystem() {
		return b.resourceRepo
	}
	return b.teamResourceRepo
}

// GetResource 获取资源详情
func (b *ResourceBiz) GetResource(ctx context.Context, id uint32) (imodel.IResource, error) {
	resource, err := b.getResourceRepo(ctx).GetByID(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastApiNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return resource, nil
}

// ListResource 获取资源列表
func (b *ResourceBiz) ListResource(ctx context.Context, params *bo.QueryResourceListParams) ([]imodel.IResource, error) {
	resourceDos, err := b.getResourceRepo(ctx).FindByPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return resourceDos, nil
}

// UpdateResourceStatus 更新资源状态
func (b *ResourceBiz) UpdateResourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	if err := b.getResourceRepo(ctx).UpdateStatus(ctx, status, ids...); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetResourceSelectList 获取资源下拉列表
func (b *ResourceBiz) GetResourceSelectList(ctx context.Context, params *bo.QueryResourceListParams) ([]*bo.SelectOptionBo, error) {
	resourceDos, err := b.getResourceRepo(ctx).FindSelectByPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}

	return types.SliceTo(resourceDos, func(resource imodel.IResource) *bo.SelectOptionBo {
		return bo.NewResourceSelectOptionBuild(resource).ToSelectOption()
	}), nil
}
