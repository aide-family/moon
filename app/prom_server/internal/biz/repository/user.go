package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
)

type (
	UserRepo interface {
		GetUserById(ctx context.Context, id uint) (*dobo.UserDO, error)
		ListUser(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.UserDO, error)
		ListUserByIds(ctx context.Context, ids []uint) ([]*dobo.UserDO, error)
		CreateUser(ctx context.Context, user *dobo.UserDO) (*dobo.UserDO, error)
		UpdateUserById(ctx context.Context, id uint, user *dobo.UserDO) (*dobo.UserDO, error)
		DeleteUserByIds(ctx context.Context, id ...uint) error
	}
)
