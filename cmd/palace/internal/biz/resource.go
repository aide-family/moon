package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/pkg/merr"
)

// NewResourceBiz 创建资源业务逻辑
func NewResourceBiz(
	resourceRepo repository.Resource,
	transaction repository.Transaction,
	logger log.Logger,
) *ResourceBiz {
	return &ResourceBiz{
		resourceRepo: resourceRepo,
		transaction:  transaction,
		helper:       log.NewHelper(log.With(logger, "module", "biz.resource")),
	}
}

type ResourceBiz struct {
	resourceRepo repository.Resource
	transaction  repository.Transaction
	helper       *log.Helper
}

func (r *ResourceBiz) BatchUpdateResourceStatus(ctx context.Context, req *bo.BatchUpdateResourceStatusReq) error {
	return r.resourceRepo.BatchUpdateResourceStatus(ctx, req.IDs, req.Status)
}

func (r *ResourceBiz) GetResource(ctx context.Context, id uint32) (do.Resource, error) {
	return r.resourceRepo.GetResourceByID(ctx, id)
}

func (r *ResourceBiz) ListResource(ctx context.Context, req *bo.ListResourceReq) (*bo.ListResourceReply, error) {
	return r.resourceRepo.ListResources(ctx, req)
}

func (r *ResourceBiz) SelfMenus(ctx context.Context) ([]do.Menu, error) {
	userID, ok := permission.GetUserIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("user id not found")
	}
	return r.resourceRepo.GetMenusByUserID(ctx, userID)
}

func (r *ResourceBiz) Menus(ctx context.Context, t vobj.MenuType) ([]do.Menu, error) {
	return r.resourceRepo.GetMenus(ctx, t)
}

func (r *ResourceBiz) SaveResource(ctx context.Context, req *bo.SaveResourceReq) error {
	return r.transaction.MainExec(ctx, func(ctx context.Context) error {
		if req.ID <= 0 {
			return r.resourceRepo.CreateResource(ctx, req)
		}
		resourceDo, err := r.resourceRepo.GetResourceByID(ctx, req.ID)
		if err != nil {
			return err
		}
		return r.resourceRepo.UpdateResource(ctx, req.WithResource(resourceDo))
	})
}

func (r *ResourceBiz) SaveMenu(ctx context.Context, req *bo.SaveMenuReq) error {
	return r.transaction.MainExec(ctx, func(ctx context.Context) error {
		if req.ParentID > 0 {
			parentDo, err := r.resourceRepo.GetMenuByID(ctx, req.ParentID)
			if err != nil {
				return err
			}
			req.WithParent(parentDo)
		}
		if len(req.ResourceIds) > 0 {
			resourceDos, err := r.resourceRepo.Find(ctx, req.ResourceIds)
			if err != nil {
				return err
			}
			req.WithResources(resourceDos)
		}
		if req.ID <= 0 {
			return r.resourceRepo.CreateMenu(ctx, req)
		}
		menuDo, err := r.resourceRepo.GetMenuByID(ctx, req.ID)
		if err != nil {
			return err
		}
		req.WithMenu(menuDo)
		return r.resourceRepo.UpdateMenu(ctx, req)
	})
}

func (r *ResourceBiz) GetMenu(ctx context.Context, id uint32) (do.Menu, error) {
	return r.resourceRepo.GetMenuByID(ctx, id)
}
