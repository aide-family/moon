package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewResourceBiz(resourceRepo repo.ResourceRepo) *ResourceBiz {
	return &ResourceBiz{
		resourceRepo: resourceRepo,
	}
}

type ResourceBiz struct {
	resourceRepo repo.ResourceRepo
}

// GetResource 获取资源详情
func (b *ResourceBiz) GetResource(ctx context.Context, id uint32) (*model.SysAPI, error) {
	resource, err := b.resourceRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bo.ResourceNotFoundErr
		}
		return nil, bo.SystemErr.WithCause(err)
	}
	return resource, nil
}

// ListResource 获取资源列表
func (b *ResourceBiz) ListResource(ctx context.Context, params *bo.QueryResourceListParams) ([]*model.SysAPI, error) {
	resourceDos, err := b.resourceRepo.FindByPage(ctx, params)
	if err != nil {
		return nil, bo.SystemErr.WithCause(err)
	}
	return resourceDos, nil
}

func (b *ResourceBiz) UpdateResourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	err := b.resourceRepo.UpdateStatus(ctx, status, ids...)
	if err != nil {
		return bo.SystemErr.WithCause(err)
	}
	return nil
}

func (b *ResourceBiz) GetResourceSelectList(ctx context.Context, params *bo.QueryResourceListParams) ([]*bo.SelectOptionBo, error) {
	resourceDos, err := b.resourceRepo.FindSelectByPage(ctx, params)
	if err != nil {
		return nil, bo.SystemErr.WithCause(err)
	}

	return types.SliceTo(resourceDos, func(resource *model.SysAPI) *bo.SelectOptionBo {
		return bo.NewResourceSelectOptionBuild(resource).ToSelectOption()
	}), nil
}
