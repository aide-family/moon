package repo

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type (
	// UserRepo is the user repository
	UserRepo interface {
		// Create a user
		Create(ctx context.Context, user *bo.CreateUserParams) (*model.SysUser, error)

		// BatchCreate batch create user
		BatchCreate(ctx context.Context, users []*bo.CreateUserParams) error

		// GetByID get user by id
		GetByID(ctx context.Context, id uint32) (*model.SysUser, error)

		// GetByUsername get user by username
		GetByUsername(ctx context.Context, username string) (*model.SysUser, error)

		// UpdateByID update user by id
		UpdateByID(ctx context.Context, user *bo.UpdateUserParams) error

		// DeleteByID delete user by id
		DeleteByID(ctx context.Context, id uint32) error

		// UpdateStatusByIds update user status by ids
		UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error

		// UpdatePassword update user password
		UpdatePassword(ctx context.Context, id uint32, password types.Password) error

		// FindByPage find user by page
		FindByPage(ctx context.Context, page *bo.QueryUserListParams) ([]*model.SysUser, error)

		// UpdateUser update user
		UpdateUser(ctx context.Context, user *model.SysUser) error

		// FindByIds find user by ids
		FindByIds(ctx context.Context, ids ...uint32) ([]*model.SysUser, error)
	}
)
