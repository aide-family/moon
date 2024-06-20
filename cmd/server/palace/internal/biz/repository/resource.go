package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

type Resource interface {
	// GetById get resource by id
	GetById(ctx context.Context, id uint32) (*model.SysAPI, error)

	// FindByPage find resource by page
	FindByPage(ctx context.Context, page *bo.QueryResourceListParams) ([]*model.SysAPI, error)

	// UpdateStatus update resource status
	UpdateStatus(ctx context.Context, status vobj.Status, ids ...uint32) error

	// FindSelectByPage find select resource by page
	FindSelectByPage(ctx context.Context, page *bo.QueryResourceListParams) ([]*model.SysAPI, error)
}
