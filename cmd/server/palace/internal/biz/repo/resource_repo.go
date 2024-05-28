package repo

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type ResourceRepo interface {
	// GetById get resource by id
	GetById(ctx context.Context, id uint32) (*model.SysAPI, error)

	// FindByPage find resource by page
	FindByPage(ctx context.Context, page *bo.QueryResourceListParams) ([]*model.SysAPI, error)

	// UpdateStatus update resource status
	UpdateStatus(ctx context.Context, status vobj.Status, ids ...uint32) error

	// FindSelectByPage find select resource by page
	FindSelectByPage(ctx context.Context, page *bo.QueryResourceListParams) ([]*model.SysAPI, error)
}
