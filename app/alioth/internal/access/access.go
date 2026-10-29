package access

import (
	"context"
	"fmt"
	"log/slog"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewChecker)

// Checker 在具体用例之前确认操作者和租户。
type Checker struct {
	Users      repository.UserRepo
	Namespaces repository.NamespaceRepo
	Roles      repository.RoleRepo
	Log        *slog.Logger
}

func NewChecker(users repository.UserRepo, namespaces repository.NamespaceRepo, roles repository.RoleRepo, logger *slog.Logger) Checker {
	return Checker{Users: users, Namespaces: namespaces, Roles: roles, Log: logger}
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

// Require 只检查权限，不返回租户。权限来自该用户在该租户的角色，页面隐藏按钮不能代替这里的拒绝。
func (c Checker) Require(ctx context.Context, code string) error {
	return c.require(ctx, code, true)
}

func (c Checker) require(ctx context.Context, code string, activeOnly bool) error {
	ns, err := c.currentNamespace(ctx, activeOnly)
	if err != nil {
		return err
	}
	uid, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	perms, err := c.Roles.Permissions(ctx, ns, uid)
	if err != nil {
		return c.Wrap(err)
	}
	if _, ok := perms[code]; ok {
		return nil
	}
	// 平台管理员可以只读切入不是自己成员的租户：只有租户目录里的 :read 放行，
	// 写、删除、介入和执行动作仍然要求该租户的成员角色。
	if perm.IsReadOnly(code) {
		admin, err := c.SystemAdmin(ctx)
		if err != nil {
			return err
		}
		if admin {
			c.Log.Debug(fmt.Sprintf("access: system admin read-only user=%s namespace=%s permission=%s", uid, ns, code))
			return nil
		}
	}
	return merr.Forbidden("permission denied")
}

// SystemAdmin 判断当前登录用户是否持有平台管理员角色。
// 只读切入只认 SUPER_ADMIN 和 SYSTEM_ADMIN；普通用户和已封禁用户都不算。
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
	return account.Status == "ACTIVE" && perm.IsSystemAdmin(account.SystemRole), nil
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
	ns, ok := actor.NamespaceID(ctx)
	if !ok {
		return "", merr.Client("X-Namespace is required")
	}
	item, err := c.Namespaces.GetNamespace(ctx, ns)
	if err != nil {
		return "", c.Wrap(err)
	}
	if activeOnly && item.Status != "ACTIVE" {
		return "", merr.Forbidden("namespace is disabled")
	}
	return ns, nil
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
