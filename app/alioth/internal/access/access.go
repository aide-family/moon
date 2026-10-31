package access

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewChecker)

// Checker 在具体用例之前确认操作者和租户，以及租户是否开通了对应功能。
type Checker struct {
	Users      repository.UserRepo
	Namespaces repository.NamespaceRepo
	Roles      repository.RoleRepo
	Features   repository.FeatureRepo
	Log        *slog.Logger
}

func NewChecker(users repository.UserRepo, namespaces repository.NamespaceRepo, roles repository.RoleRepo, features repository.FeatureRepo, logger *slog.Logger) Checker {
	return Checker{Users: users, Namespaces: namespaces, Roles: roles, Features: features, Log: logger}
}

// RequireSystem 校验平台权限。它不读取 X-Namespace，也不依赖租户成员身份。
// 用户列表、封禁和系统角色授权只能走这里。
func (c Checker) RequireSystem(ctx context.Context, code string) (*bo.Account, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := c.Users.UserByUID(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.Unauthenticated("user is not active")
		}
		return nil, c.Wrap(err)
	}
	if account.Status != "ACTIVE" {
		return nil, merr.Unauthenticated("user is not active")
	}
	for _, name := range perm.SystemPermissions(account.SystemRole) {
		if name == code {
			return account, nil
		}
	}
	return nil, merr.Forbidden("permission denied")
}

// SystemPermissions 返回当前用户的平台权限，供 /self/info 使用。
func (c Checker) SystemPermissions(ctx context.Context) ([]string, string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, "", merr.Unauthenticated("missing token")
	}
	account, err := c.Users.UserByUID(ctx, uid)
	if err != nil {
		return nil, "", c.Wrap(err)
	}
	return perm.SystemPermissions(account.SystemRole), account.SystemRole, nil
}

// Allow 确认当前用户持有 code，并返回当前租户。
// 缺少 X-Namespace、租户不存在、租户已停用、用户未登录或没有该权限时，调用方必须停止写操作。
func (c Checker) Allow(ctx context.Context, code string) (string, error) {
	return c.allow(ctx, code, true)
}

// AllowAnyStatus 与 Allow 相同，但允许在已停用的租户上执行。
// 只用于租户自身的查看、启停和删除，避免停用后无法再启用。
func (c Checker) AllowAnyStatus(ctx context.Context, code string) (string, error) {
	return c.allow(ctx, code, false)
}

func (c Checker) allow(ctx context.Context, code string, activeOnly bool) (string, error) {
	if err := c.require(ctx, code, activeOnly); err != nil {
		return "", err
	}
	ns, _ := actor.NamespaceID(ctx)
	return ns, nil
}

// RequireAny 只要命中任意一个权限就放行；用于同时服务多个菜单的接口，
// 例如实时告警和历史告警共用的事件列表。返回最后一个错误，便于定位。
func (c Checker) RequireAny(ctx context.Context, codes ...string) error {
	_, err := c.requireAny(ctx, codes...)
	return err
}

// AllowAny 与 Allow 相同，但允许传入多个权限，命中任意一个即返回租户。
func (c Checker) AllowAny(ctx context.Context, codes ...string) (string, error) {
	return c.requireAny(ctx, codes...)
}

func (c Checker) requireAny(ctx context.Context, codes ...string) (string, error) {
	var last error
	for _, code := range codes {
		ns, err := c.allow(ctx, code, true)
		if err == nil {
			return ns, nil
		}
		last = err
	}
	if last == nil {
		return "", merr.Forbidden("permission denied")
	}
	return "", last
}

// Require 只检查权限，不返回租户。权限来自该用户在该租户的角色，页面隐藏按钮不能代替这里的拒绝。
func (c Checker) Require(ctx context.Context, code string) error {
	return c.require(ctx, code, true)
}

func (c Checker) require(ctx context.Context, code string, activeOnly bool) error {
	ns, namespace, err := c.currentNamespaceItem(ctx, activeOnly)
	if err != nil {
		return err
	}
	uid, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	// 租户负责人隐式持有全部租户权限，不依赖任何预置角色数据。
	// 套餐能力/模块检查仍然生效，负责人也不能使用租户未开通的功能。
	owner := namespace.OwnerUID == uid
	var perms map[string]struct{}
	if !owner {
		perms, err = c.Roles.Permissions(ctx, ns, uid)
		if err != nil {
			return c.Wrap(err)
		}
	}
	readOnly := false
	if !owner {
		if _, ok := perms[code]; !ok {
			// 平台管理员可以只读切入不是自己成员的租户：只有租户目录里的 :read 放行，
			// 写、删除、介入和执行动作仍然要求该租户的成员角色。
			// 系统只读角色同样走这条只读通道，它没有任何写权限。
			if !perm.IsReadOnly(code) {
				return merr.Forbidden("permission denied")
			}
			admin, err := c.SystemAdmin(ctx)
			if err != nil {
				return err
			}
			if !admin {
				return merr.Forbidden("permission denied")
			}
			readOnly = true
			c.Log.Debug(fmt.Sprintf("access: system admin read-only user=%s namespace=%s permission=%s", uid, ns, code))
		}
	}
	// namespace:* 不属于任何能力，始终可用；菜单权限本身不是接口权限，不在这里判断。
	module := perm.FeatureOf(code)
	if module == "" || perm.IsMenuPermission(code) {
		return nil
	}
	// 权限通过后再看租户开通情况：能力没开通，成员持有权限也不能用。
	// 平台管理员的只读视角同样受限。
	modules, err := c.Features.EffectiveModules(ctx, ns)
	if err != nil {
		return c.Wrap(err)
	}
	if !slices.Contains(modules, module) {
		return merr.Forbidden("feature not enabled")
	}
	// 负责人的权限集合就是整个租户权限目录（含菜单权限），菜单检查必然通过。
	if owner {
		return nil
	}
	// 跨租户只读的平台管理员不需要租户给它授菜单权限，菜单只约束租户成员。
	if readOnly {
		return nil
	}
	// 菜单权限决定租户成员能不能打开对应页面：同一个能力下只要有一条 menu:<cap>
	// 就放行。角色授权页由 ListPermissions 只展示已开通能力，两者范围一致。
	capabilities, err := c.Features.EffectiveCapabilities(ctx, ns)
	if err != nil {
		return c.Wrap(err)
	}
	if c.hasMenuPermission(perms, capabilities, code) {
		return nil
	}
	return merr.Forbidden("menu permission denied")
}

// HeldPermissions 返回当前用户在当前租户持有的权限集合，第二个返回值表示是否为租户负责人。
// 负责人隐式持有全部租户权限（perm.Names()，含菜单权限），其余用户取角色权限的并集。
// 写授权前用它做「不能授予自己没有的权限」的子集校验。
// 它不校验租户状态和成员身份，需要鉴权的入口必须先走 Allow / Require。
func (c Checker) HeldPermissions(ctx context.Context) (map[string]struct{}, bool, error) {
	ns, ok := actor.NamespaceID(ctx)
	if !ok {
		return nil, false, merr.Client("X-Namespace is required")
	}
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, false, merr.Unauthenticated("missing token")
	}
	namespace, err := c.Namespaces.GetNamespace(ctx, ns)
	if err != nil {
		return nil, false, c.Wrap(err)
	}
	if namespace.OwnerUID == uid {
		return fullTenantPermissions(), true, nil
	}
	perms, err := c.Roles.Permissions(ctx, ns, uid)
	if err != nil {
		return nil, false, c.Wrap(err)
	}
	return perms, false, nil
}

// fullTenantPermissions 把租户权限目录转成集合。负责人隐式持有这份完整目录。
func fullTenantPermissions() map[string]struct{} {
	names := perm.Names()
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}

// Capabilities 返回当前租户当前生效的能力。租户停用、归档或套餐到期时为空。
// 数据源、探测和通知渠道按类型过滤资源时用它决定哪些类型可见。
func (c Checker) Capabilities(ctx context.Context) ([]string, error) {
	ns, err := c.currentNamespace(ctx, true)
	if err != nil {
		return nil, err
	}
	capabilities, err := c.Features.EffectiveCapabilities(ctx, ns)
	if err != nil {
		return nil, c.Wrap(err)
	}
	return capabilities, nil
}

// RequireCapability 要求当前租户已开通某个能力。它用于资源类型级检查：
// 模块权限通过之后，具体类型（某个数据源驱动 / 探测驱动 / 通知渠道类型）
// 还要在套餐范围内，否则返回 403 feature not enabled。
func (c Checker) RequireCapability(ctx context.Context, capability string) error {
	capabilities, err := c.Capabilities(ctx)
	if err != nil {
		return err
	}
	if !slices.Contains(capabilities, capability) {
		return merr.Forbidden("feature not enabled")
	}
	return nil
}

// hasMenuPermission 判断角色是否持有能覆盖 code 的菜单权限。
// 能力必须已在租户开通范围内，且角色的权限集合里出现对应的 menu:<能力>。
func (c Checker) hasMenuPermission(perms map[string]struct{}, capabilities []string, code string) bool {
	// 权限声明的菜单中，只要有一个被租户能力开通、且角色持有它的菜单权限，即可入口。
	for _, menu := range perm.MenusOfPermission(code) {
		if !perm.MenuLicensed(menu, capabilities) {
			continue
		}
		if _, held := perms[perm.MenuPermission(menu)]; held {
			return true
		}
	}
	return false
}

// SystemAdmin 判断当前登录用户是否可以跨租户只读查看。
// SUPER_ADMIN、SYSTEM_ADMIN 和 SYSTEM_READONLY 都算；普通用户和已封禁用户都不算。
// 只是只读通道，写操作仍然要求租户成员角色。
func (c Checker) SystemAdmin(ctx context.Context) (bool, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return false, merr.Unauthenticated("missing token")
	}
	account, err := c.Users.UserByUID(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return false, nil
		}
		return false, c.Wrap(err)
	}
	return account.Status == "ACTIVE" && perm.CanReadAnyTenant(account.SystemRole), nil
}

// CurrentNamespace 返回 X-Namespace 指定的租户，不校验成员身份和租户状态。
// /self/permissions 用它判断租户是否停用或已归档；没有租户头或租户已删除时返回 nil。
func (c Checker) CurrentNamespace(ctx context.Context) (*bo.Namespace, error) {
	ns, ok := actor.NamespaceID(ctx)
	if !ok {
		return nil, nil
	}
	item, err := c.Namespaces.GetNamespace(ctx, ns)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, nil
		}
		return nil, c.Wrap(err)
	}
	return item, nil
}

// SameNamespace 要求路径中的租户就是当前租户，避免用一个租户的令牌改另一个租户。
func (c Checker) SameNamespace(ctx context.Context, uid string) error {
	return c.sameNamespace(ctx, uid, true)
}

// SameNamespaceAnyStatus 与 SameNamespace 相同，但允许已停用的租户，供查看、启停和删除使用。
func (c Checker) SameNamespaceAnyStatus(ctx context.Context, uid string) error {
	return c.sameNamespace(ctx, uid, false)
}

func (c Checker) sameNamespace(ctx context.Context, uid string, activeOnly bool) error {
	ns, err := c.currentNamespace(ctx, activeOnly)
	if err != nil {
		return err
	}
	if ns != uid {
		return merr.Client("namespace mismatch")
	}
	return nil
}

func (c Checker) currentNamespace(ctx context.Context, activeOnly bool) (string, error) {
	ns, _, err := c.currentNamespaceItem(ctx, activeOnly)
	return ns, err
}

// currentNamespaceItem 同时返回租户 UID 和租户行：负责人判断需要 owner_uid。
func (c Checker) currentNamespaceItem(ctx context.Context, activeOnly bool) (string, *bo.Namespace, error) {
	ns, ok := actor.NamespaceID(ctx)
	if !ok {
		return "", nil, merr.Client("X-Namespace is required")
	}
	item, err := c.Namespaces.GetNamespace(ctx, ns)
	if err != nil {
		return "", nil, c.Wrap(err)
	}
	if activeOnly && item.Status != "ACTIVE" {
		return "", nil, merr.Forbidden("namespace is disabled")
	}
	return ns, item, nil
}

// Wrap 直接返回仓储给出的 Kratos error。意外的原始错误只记日志，不把数据库细节返回去。
func (c Checker) Wrap(err error) error {
	if err == nil {
		return nil
	}
	if kerrors.Reason(err) != "" {
		return err
	}
	c.Log.Error(fmt.Sprintf("biz: %v", err))
	return merr.Server("request failed")
}
