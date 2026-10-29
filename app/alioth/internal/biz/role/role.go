package role

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewAuthority)

// Authority 管理角色和角色的权限授予。
type Authority struct {
	roles  repository.RoleRepo
	access access.Checker
}

func NewAuthority(roles repository.RoleRepo, checker access.Checker) *Authority {
	return &Authority{roles: roles, access: checker}
}

func (authority *Authority) CreateRole(ctx context.Context, name, remark string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := authority.roles.CreateRole(ctx, ns, name, remark)
	return item, authority.access.Wrap(err)
}

// UpdateRoleRemark 只改备注，系统角色也能改：备注是给人看的说明，不改变预置的权限边界。
// 不要把 IsSystem 判断照抄进来，那会让系统角色失去唯一能解释自己的字段。
func (authority *Authority) UpdateRoleRemark(ctx context.Context, uid, remark string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := authority.roles.UpdateRoleRemark(ctx, ns, uid, remark)
	return item, authority.access.Wrap(err)
}

// UpdateRole 只改自定义角色的名称。owner、admin、member、viewer 是系统角色，改名会破坏预置边界。
func (authority *Authority) UpdateRole(ctx context.Context, uid, name string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := authority.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, authority.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	item, err := authority.roles.UpdateRole(ctx, ns, uid, name)
	return item, authority.access.Wrap(err)
}

func (authority *Authority) UpdateRoleStatus(ctx context.Context, uid, status string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := authority.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, authority.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	item, err := authority.roles.UpdateRoleStatus(ctx, ns, uid, status)
	return item, authority.access.Wrap(err)
}

// DeleteRole 不删除系统角色。删掉 owner 会让租户失去负责人角色。
func (authority *Authority) DeleteRole(ctx context.Context, uid string) error {
	ns, err := authority.access.Allow(ctx, perm.RoleDelete)
	if err != nil {
		return err
	}
	current, err := authority.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return authority.access.Wrap(err)
	}
	if current.IsSystem {
		return merr.Client("invalid request")
	}
	return authority.access.Wrap(authority.roles.DeleteRole(ctx, ns, uid))
}

func (authority *Authority) GetRole(ctx context.Context, uid string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	item, err := authority.roles.GetRole(ctx, ns, uid)
	return item, authority.access.Wrap(err)
}

func (authority *Authority) ListRoles(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Role, int, int, int, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := authority.roles.ListRoles(ctx, ns, p, s, name)
	return items, total, p, s, authority.access.Wrap(err)
}

func (authority *Authority) SelectRoles(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	items, err := authority.roles.SelectRoles(ctx, ns, page.Limit(limit), lastUID)
	return items, authority.access.Wrap(err)
}

// UpdateRolePermissions 替换自定义角色的权限。
// namespace:delete 和 namespace:transfer 只属于 owner，不能授给自定义角色。
// 系统角色的权限由预置目录决定，这里直接拒绝。
func (authority *Authority) UpdateRolePermissions(ctx context.Context, uid string, names []string) (*bo.Role, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := authority.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, authority.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	for _, name := range names {
		if !perm.IsTenantPermission(name) || name == perm.NamespaceDelete || name == perm.NamespaceTransfer {
			return nil, merr.Client("invalid request")
		}
	}
	item, err := authority.roles.UpdateRolePermissions(ctx, ns, uid, names)
	return item, authority.access.Wrap(err)
}

func (authority *Authority) ListPermissions(ctx context.Context) ([]bo.Permission, error) {
	if _, err := authority.access.Allow(ctx, perm.RoleRead); err != nil {
		return nil, err
	}
	items, err := authority.roles.ListPermissions(ctx)
	return items, authority.access.Wrap(err)
}

// Mine 返回当前用户在当前租户的权限名，以及是否只是平台管理员的只读视角。
// 没有选择租户时返回空列表和 false，而不是报错。
// 不是该租户成员、但持有平台管理员角色时，返回租户目录中的全部只读权限，
// 调用方据此展示只读横幅；写操作仍由 access 按成员角色拒绝。
func (authority *Authority) Mine(ctx context.Context) ([]string, bool, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, false, merr.Unauthenticated("missing token")
	}
	namespace, ok := actor.NamespaceID(ctx)
	if !ok {
		return []string{}, false, nil
	}
	perms, err := authority.roles.Permissions(ctx, namespace, uid)
	if err != nil {
		return nil, false, authority.access.Wrap(err)
	}
	if len(perms) == 0 {
		// 权限为空时区分两种情况：成员只是没有有效权限，仍然按成员返回空列表；
		// 非成员的平台管理员才能拿到只读目录。
		member, err := authority.roles.IsMember(ctx, namespace, uid)
		if err != nil {
			return nil, false, authority.access.Wrap(err)
		}
		if !member {
			admin, err := authority.access.SystemAdmin(ctx)
			if err != nil {
				return nil, false, err
			}
			if admin {
				return perm.ReadOnlyNames(), true, nil
			}
		}
		return []string{}, false, nil
	}
	names := make([]string, 0, len(perms))
	for name := range perms {
		names = append(names, name)
	}
	return names, false, nil
}
