package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

// Resource 资源管理接口
type Resource interface {
	// GetById get resource by id
	GetByID(context.Context, uint32) (*model.SysAPI, error)

	// FindByPage find resource by page
	FindByPage(context.Context, *bo.QueryResourceListParams) ([]*model.SysAPI, error)

	// UpdateStatus update resource status
	UpdateStatus(context.Context, vobj.Status, ...uint32) error

	// FindSelectByPage find select resource by page
	FindSelectByPage(context.Context, *bo.QueryResourceListParams) ([]*model.SysAPI, error)
}
