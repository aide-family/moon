package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	UserRepo interface {
		Get(ctx context.Context, scopes ...query.ScopeMethod) (*dobo.UserDO, error)
		List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.UserDO, error)
		Create(ctx context.Context, user *dobo.UserDO) (*dobo.UserDO, error)
		Update(ctx context.Context, user *dobo.UserDO, scopes ...query.ScopeMethod) (*dobo.UserDO, error)
		Delete(ctx context.Context, scopes ...query.ScopeMethod) error
	}
)
