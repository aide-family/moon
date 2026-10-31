package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
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
	// MergeAccounts 把 fromUID 的租户成员、租户角色、租户负责人、平台系统角色和全部个人数据
	// 并入 toUID，并在同一事务里软删源账号。目标账号是母本：目标已有冲突行时保留目标、丢弃源行。
	// 第三方占位账号补邮箱时邮箱已被另一个账号占用，会调用它合并账号。
	MergeAccounts(ctx context.Context, fromUID, toUID string) error
	// DeleteUser 软删用户并清理其第三方绑定、成员与成员角色、告警页个人选择、告警列配置、
	// 系统消息和邀请记录；保留其创建过的租户告警页面。用户仍是租户负责人或还有租户订单时拒绝。
	DeleteUser(ctx context.Context, uid string) error
}
