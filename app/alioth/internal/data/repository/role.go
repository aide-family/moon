package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// RoleRepo 读写租户角色，并计算某个用户在租户内的有效权限。
type RoleRepo interface {
	ListRoles(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Role, int, error)
	GetRole(ctx context.Context, namespaceUID, uid string) (*bo.Role, error)
	SelectRoles(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error)
	CreateRole(ctx context.Context, namespaceUID, name, remark string) (*bo.Role, error)
	UpdateRole(ctx context.Context, namespaceUID, uid, name string) (*bo.Role, error)
	// UpdateRoleRemark 只改备注。系统角色的备注也能改，它不改变预置的权限边界。
	UpdateRoleRemark(ctx context.Context, namespaceUID, uid, remark string) (*bo.Role, error)
	UpdateRoleStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Role, error)
	DeleteRole(ctx context.Context, namespaceUID, uid string) error
	UpdateRolePermissions(ctx context.Context, namespaceUID, uid string, names []string) (*bo.Role, error)
	ListPermissions(ctx context.Context) ([]bo.Permission, error)
	// IsMember 判断用户是否是租户的 JOINED 成员，用来区分「成员但没有有效权限」和「非成员」。
	IsMember(ctx context.Context, namespaceUID, userUID string) (bool, error)
	Permissions(ctx context.Context, namespaceUID, userUID string) (map[string]struct{}, error)
}
