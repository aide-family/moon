package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Resource interface {
	// GetResources get all resources
	GetResources(ctx context.Context) ([]do.Resource, error)

	// GetResourceByID get resource by id
	GetResourceByID(ctx context.Context, id uint32) (do.Resource, error)

	// GetResourceByOperation get resource by operation
	GetResourceByOperation(ctx context.Context, operation string) (do.Resource, error)

	// ListResources list resources
	ListResources(ctx context.Context, req *bo.ListResourceReq) (*bo.ListResourceReply, error)

	// SelectResources select resources
	SelectResources(ctx context.Context, req *bo.SelectResourceReq) (*bo.SelectResourceReply, error)

	// BatchUpdateResourceStatus update multiple resources status
	BatchUpdateResourceStatus(ctx context.Context, ids []uint32, status vobj.GlobalStatus) error

	// GetMenusByUserID get all menus
	GetMenusByUserID(ctx context.Context, userID uint32) ([]do.Menu, error)

	// GetMenus get all menus
	GetMenus(ctx context.Context, t vobj.MenuType) ([]do.Menu, error)

	// CreateResource create resource
	CreateResource(ctx context.Context, req bo.SaveResource) error

	// UpdateResource update resource
	UpdateResource(ctx context.Context, req bo.SaveResource) error

	// CreateMenu create menu
	CreateMenu(ctx context.Context, req bo.SaveMenu) error

	// UpdateMenu update menu
	UpdateMenu(ctx context.Context, req bo.SaveMenu) error

	// GetMenuByID get menu by id
	GetMenuByID(ctx context.Context, id uint32) (do.Menu, error)

	// Find find resources by ids
	Find(ctx context.Context, ids []uint32) ([]do.Resource, error)
}
