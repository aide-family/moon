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
	UpdateUserProfile(ctx context.Context, uid string, email, phone, avatar, remark *string) (*bo.Account, error)
	UpdateUserStatus(ctx context.Context, uid, status string) (*bo.Account, error)
	// UpdateUserSystemRole 只改平台角色，并记录授权人。超级管理员唯一性由数据库索引保证。
	UpdateUserSystemRole(ctx context.Context, uid, role, authorizedBy string) (*bo.Account, error)
	ListUsers(ctx context.Context, page, pageSize int, name string) ([]bo.Account, int, error)
	SelectUsers(ctx context.Context, limit int, lastUID string) ([]bo.SelectItem, error)
	// MergeAccounts 把 fromUID 的租户成员、租户角色和平台系统角色迁移到 toUID。
	// 第三方占位账号补邮箱时邮箱已被另一个账号占用，会调用它合并账号。
	// 个人数据（列表列、告警页选择、系统消息等）不迁移，源账号保留为空账号。
	MergeAccounts(ctx context.Context, fromUID, toUID string) error
}
