package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// UserRepo 读写全局用户。
type UserRepo interface {
	CreateUser(ctx context.Context, in bo.Account) (*bo.Account, error)
	UserByEmail(ctx context.Context, email string) (*bo.Account, error)
	UserByUID(ctx context.Context, uid string) (*bo.Account, error)
	UpdateUserProfile(ctx context.Context, uid string, email, avatar, remark *string) (*bo.Account, error)
	UpdateUserStatus(ctx context.Context, uid, status string) (*bo.Account, error)
	ListUsers(ctx context.Context, page, pageSize int, name string) ([]bo.Account, int, error)
	SelectUsers(ctx context.Context, limit int, lastUID string) ([]bo.SelectItem, error)
}
