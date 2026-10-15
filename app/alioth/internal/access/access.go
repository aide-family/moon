package access

import (
	"context"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewChecker)

// Checker 在具体用例之前确认操作者和租户。
type Checker struct {
	Namespaces repository.NamespaceRepo
	Roles      repository.RoleRepo
	Log        *log.Helper
}

func NewChecker(namespaces repository.NamespaceRepo, roles repository.RoleRepo, logger log.Logger) Checker {
	return Checker{Namespaces: namespaces, Roles: roles, Log: log.NewHelper(logger)}
}

// Allow 确认当前用户持有 code，并返回当前租户。
// 缺少 X-Namespace、租户不存在、用户未登录或没有该权限时，调用方必须停止写操作。
func (c Checker) Allow(ctx context.Context, code string) (string, error) {
	if err := c.Require(ctx, code); err != nil {
		return "", err
	}
	ns, _ := actor.NamespaceID(ctx)
	return ns, nil
}

// Require 只检查权限，不返回租户。权限来自该用户在该租户的角色，页面隐藏按钮不能代替这里的拒绝。
func (c Checker) Require(ctx context.Context, code string) error {
	ns, err := c.currentNamespace(ctx)
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
	if _, ok := perms[code]; !ok {
		return merr.Forbidden("permission denied")
	}
	return nil
}

// SameNamespace 要求路径中的租户就是当前租户，避免用一个租户的令牌改另一个租户。
func (c Checker) SameNamespace(ctx context.Context, uid string) error {
	ns, err := c.currentNamespace(ctx)
	if err != nil {
		return err
	}
	if ns != uid {
		return merr.Client("namespace mismatch")
	}
	return nil
}

func (c Checker) currentNamespace(ctx context.Context) (string, error) {
	ns, ok := actor.NamespaceID(ctx)
	if !ok {
		return "", merr.Client("X-Namespace is required")
	}
	if _, err := c.Namespaces.GetNamespace(ctx, ns); err != nil {
		return "", c.Wrap(err)
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
	c.Log.Errorf("usecase: %v", err)
	return merr.Server("request failed")
}
