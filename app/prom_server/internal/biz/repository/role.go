package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	RoleRepo interface {
		CreateRole(ctx context.Context, role *dobo.RoleDO) (*dobo.RoleDO, error)
		UpdateRoleById(ctx context.Context, id uint, role *dobo.RoleDO) (*dobo.RoleDO, error)
		DeleteRoleByIds(ctx context.Context, id ...uint) error
		GetRoleById(ctx context.Context, id uint) (*dobo.RoleDO, error)
		ListRole(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.RoleDO, error)
	}
)
