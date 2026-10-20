package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// UserRepo 读写平台级用户。用户不属于任何租户。
type UserRepo interface {
	// CreateUser 创建用户。平台还没有用户时，第一个用户自动成为超级管理员。
	CreateUser(ctx context.Context, in bo.Account) (*bo.Account, error)
	UserByEmail(ctx context.Context, email string) (*bo.Account, error)
	UserByUID(ctx context.Context, uid string) (*bo.Account, error)
	UpdateUserProfile(ctx context.Context, uid string, email, avatar, remark *string) (*bo.Account, error)
	UpdateUserStatus(ctx context.Context, uid, status string) (*bo.Account, error)
	// UpdateUserSystemRole 只改平台角色，并记录授权人。超级管理员唯一性由数据库索引保证。
	UpdateUserSystemRole(ctx context.Context, uid, role, authorizedBy string) (*bo.Account, error)
	ListUsers(ctx context.Context, page, pageSize int, name string) ([]bo.Account, int, error)
	SelectUsers(ctx context.Context, limit int, lastUID string) ([]bo.SelectItem, error)
}
