package role

import (
	"context"
	"time"

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

// Authority 管理角色和角色的权限授予。租户只能基于已开通功能授权。
type Authority struct {
	roles    repository.RoleRepo
	features repository.FeatureRepo
	access   access.Checker
}

func NewAuthority(roles repository.RoleRepo, features repository.FeatureRepo, checker access.Checker) *Authority {
	return &Authority{roles: roles, features: features, access: checker}
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
// 只允许当前租户开通范围内的权限：模块权限按有效模块，菜单权限按有效能力。
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
	capabilities, modules, err := authority.tenantScope(ctx, ns)
	if err != nil {
		return nil, err
	}
	// 已关闭能力/模块的旧授权可以保留：编辑角色时前端会把当前权限原样回传，
	// 一律拒绝会让保存失败，静默丢弃又会丢数据。只拒绝新增的未开通权限。
	held := make(map[string]struct{}, len(current.Permissions))
	for _, name := range current.Permissions {
		held[name] = struct{}{}
	}
	for _, name := range names {
		if !perm.IsTenantPermission(name) || name == perm.NamespaceDelete || name == perm.NamespaceTransfer {
			return nil, merr.Client("invalid request")
		}
		// 未开通能力/模块的权限不能新授给角色：租户管理员不能绕过套餐自己加开。
		if !permissionInTenantScope(name, capabilities, modules) {
			if _, ok := held[name]; !ok {
				return nil, merr.Client("feature not enabled")
			}
		}
	}
	// 菜单权限带默认读权限，子权限必须挂在某个已开通的菜单下：
	// 前端已按这个规则提交，这里再归一化一次，避免直接调接口绕过。
	names = normalizeRoleGrant(names, capabilities, modules)
	item, err := authority.roles.UpdateRolePermissions(ctx, ns, uid, names)
	return item, authority.access.Wrap(err)
}

// normalizeRoleGrant 给菜单补默认读权限，并给子权限补它所属的菜单权限。
// 只有当前租户开通范围内的能力会补菜单；已关闭能力的旧授权保持原样。
func normalizeRoleGrant(names, capabilities, modules []string) []string {
	seen := make(map[string]struct{}, len(names)*2)
	out := make([]string, 0, len(names)*2)
	add := func(name string) {
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, name := range names {
		add(name)
	}
	// 先给子权限补所属菜单，再统一按菜单补默认读权限：两个方向的补全都落到最终集合。
	for _, name := range append([]string(nil), out...) {
		if perm.IsMenuPermission(name) || perm.FeatureOf(name) == "" {
			continue
		}
		for _, menu := range perm.MenusOfPermission(name) {
			if !perm.MenuLicensed(menu, capabilities) {
				continue
			}
			add(perm.MenuPermission(menu))
			break
		}
	}
	// 有菜单权限就默认带读权限（:read / :logs）。
	for _, name := range append([]string(nil), out...) {
		capability := perm.MenuPermissionCapability(name)
		if capability == "" {
			continue
		}
		for _, read := range perm.DefaultReadsOfMenu(capability) {
			if permissionInTenantScope(read, capabilities, modules) {
				add(read)
			}
		}
	}
	return out
}

// ListPermissions 只返回当前租户开通范围内的权限：模块权限按有效模块，
// 菜单权限按有效能力（menu:<能力>）。授权页看不到未开通的能力。
func (authority *Authority) ListPermissions(ctx context.Context) ([]bo.Permission, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	capabilities, modules, err := authority.tenantScope(ctx, ns)
	if err != nil {
		return nil, err
	}
	items, err := authority.roles.ListPermissions(ctx)
	if err != nil {
		return nil, authority.access.Wrap(err)
	}
	out := make([]bo.Permission, 0, len(items))
	for _, item := range items {
		if !permissionInTenantScope(item.Name, capabilities, modules) {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// ListPermissionMenus 返回角色授权页需要的「菜单 → 子权限」目录。
// 菜单只在租户已开通能力范围内返回；Always 菜单（租户设置）始终返回。
// 共享权限会出现在多个菜单下，由前端按动作去重展示。
func (authority *Authority) ListPermissionMenus(ctx context.Context) ([]bo.Permission, []bo.PermissionMenu, error) {
	ns, err := authority.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, nil, err
	}
	capabilities, modules, err := authority.tenantScope(ctx, ns)
	if err != nil {
		return nil, nil, err
	}
	items, err := authority.roles.ListPermissions(ctx)
	if err != nil {
		return nil, nil, authority.access.Wrap(err)
	}
	out := make([]bo.Permission, 0, len(items))
	byName := make(map[string]bo.Permission, len(items))
	for _, item := range items {
		if !permissionInTenantScope(item.Name, capabilities, modules) {
			continue
		}
		out = append(out, item)
		byName[item.Name] = item
	}
	menus := make([]bo.PermissionMenu, 0, len(perm.Menus()))
	for _, menu := range perm.Menus() {
		if !perm.MenuLicensed(menu.Code, capabilities) {
			continue
		}
		permissions := make([]bo.Permission, 0, len(menu.Permissions))
		for _, name := range menu.Permissions {
			if item, ok := byName[name]; ok {
				permissions = append(permissions, item)
			}
		}
		menus = append(menus, bo.PermissionMenu{
			Code: menu.Code, Group: menu.Group, Name: menu.Name,
			Always: menu.Always, Capabilities: menu.Capabilities,
			DefaultReads: menu.DefaultReads, Permissions: permissions,
		})
	}
	return out, menus, nil
}

// Mine 返回当前用户在当前租户的权限名、有效能力和是否只是平台管理员的只读视角。
// 没有选择租户时返回空列表和 false，而不是报错。
// 不是该租户成员、但可以跨租户只读的平台角色，返回只读目录并附上菜单权限，
// 调用方据此展示只读横幅；写操作仍由 access 按成员角色拒绝。
// 租户停用或已归档时返回空权限和能力，不报错；平台管理员只读切入停用租户被拒绝。
func (authority *Authority) Mine(ctx context.Context) ([]string, []string, bool, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, nil, false, merr.Unauthenticated("missing token")
	}
	namespace, ok := actor.NamespaceID(ctx)
	if !ok {
		return []string{}, []string{}, false, nil
	}
	current, err := authority.access.CurrentNamespace(ctx)
	if err != nil {
		return nil, nil, false, err
	}
	// 套餐已过期时租户仍可能是 ACTIVE，但能力已经失效，等同停用处理。
	expired := current != nil && !current.PackageExpiresAt.IsZero() && !current.PackageExpiresAt.After(time.Now())
	if current == nil || current.Status != "ACTIVE" || expired {
		// 归档租户查不到；停用租户仍在库里。成员只看到空权限，
		// 平台管理员的只读切入则直接拒绝，避免把停用租户当成可读目标。
		if current != nil && current.Status != "ACTIVE" {
			member, err := authority.roles.IsMember(ctx, namespace, uid)
			if err != nil {
				return nil, nil, false, authority.access.Wrap(err)
			}
			if !member {
				admin, err := authority.access.SystemAdmin(ctx)
				if err != nil {
					return nil, nil, false, err
				}
				if admin {
					return nil, nil, false, merr.Forbidden("namespace is disabled")
				}
			}
		}
		return []string{}, []string{}, false, nil
	}
	capabilities, modules, err := authority.tenantScope(ctx, namespace)
	if err != nil {
		return nil, nil, false, err
	}
	perms, err := authority.roles.Permissions(ctx, namespace, uid)
	if err != nil {
		return nil, nil, false, authority.access.Wrap(err)
	}
	if len(perms) == 0 {
		// 权限为空时区分两种情况：成员只是没有有效权限，仍然按成员返回空列表；
		// 非成员且可以跨租户只读的平台角色才能拿到只读目录。
		member, err := authority.roles.IsMember(ctx, namespace, uid)
		if err != nil {
			return nil, nil, false, authority.access.Wrap(err)
		}
		if !member {
			admin, err := authority.access.SystemAdmin(ctx)
			if err != nil {
				return nil, nil, false, err
			}
			if admin {
				return readOnlyPermissions(capabilities, modules), capabilities, true, nil
			}
		}
		return []string{}, capabilities, false, nil
	}
	names := make([]string, 0, len(perms))
	for name := range perms {
		names = append(names, name)
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if !permissionInTenantScope(name, capabilities, modules) {
			continue
		}
		out = append(out, name)
	}
	return out, capabilities, false, nil
}

// tenantScope 返回租户当前开通的能力和模块，两者都由套餐与租户差异计算。
func (authority *Authority) tenantScope(ctx context.Context, namespaceUID string) (capabilities, modules []string, err error) {
	capabilities, err = authority.features.EffectiveCapabilities(ctx, namespaceUID)
	if err != nil {
		return nil, nil, authority.access.Wrap(err)
	}
	modules, err = authority.features.EffectiveModules(ctx, namespaceUID)
	if err != nil {
		return nil, nil, authority.access.Wrap(err)
	}
	return capabilities, modules, nil
}

// permissionInTenantScope 判断一条租户权限是否在当前开通范围内。
// 菜单权限看对应能力；普通权限看所属模块；namespace:* 不属于任何能力，始终可用。
func permissionInTenantScope(name string, capabilities, modules []string) bool {
	if perm.IsMenuPermission(name) {
		menu, ok := perm.MenuOfMenuPermission(name)
		return ok && perm.MenuLicensed(menu, capabilities)
	}
	return perm.PermissionLicensed(name, capabilities)
}

// readOnlyPermissions 返回平台管理员只读切入时的权限目录：
// 只读权限按有效模块过滤，并补上每个已开通能力的菜单权限，保证菜单可见。
func readOnlyPermissions(capabilities, modules []string) []string {
	seen := make(map[string]struct{}, len(capabilities)*2)
	out := make([]string, 0, len(capabilities)*2)
	for _, name := range perm.ReadOnlyNames() {
		if !permissionInTenantScope(name, capabilities, modules) {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, menu := range perm.EffectiveMenus(capabilities) {
		name := perm.MenuPermission(menu)
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}
