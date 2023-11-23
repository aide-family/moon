package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	RoleRepo interface {
		Create(ctx context.Context, role *dobo.RoleDO) (*dobo.RoleDO, error)
		Update(ctx context.Context, role *dobo.RoleDO, scopes ...query.ScopeMethod) (*dobo.RoleDO, error)
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.RoleDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.RoleDO, error)
	}
)
