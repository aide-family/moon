package access

import (
	"context"
	"log/slog"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// 测试替身只实现 require 会用到的方法，其余方法由嵌入的接口提供；
// 如果实现里漏调权限检查、走到别的仓储方法，会直接 panic 而不是静默通过。
type userRepoStub struct {
	repository.UserRepo
	account *bo.Account
}

func (s userRepoStub) UserByUID(context.Context, string) (*bo.Account, error) {
	if s.account == nil {
		return nil, merr.NotFound("not found")
	}
	return s.account, nil
}

type namespaceRepoStub struct {
	repository.NamespaceRepo
	status string
}

func (s namespaceRepoStub) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	status := s.status
	if status == "" {
		status = "ACTIVE"
	}
	return &bo.Namespace{UID: uid, Status: status}, nil
}

type roleRepoStub struct {
	repository.RoleRepo
	permissions map[string]struct{}
}

func (s roleRepoStub) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return s.permissions, nil
}

func checkerFor(systemRole string, permissions ...string) Checker {
	granted := make(map[string]struct{}, len(permissions))
	for _, name := range permissions {
		granted[name] = struct{}{}
	}
	return Checker{
		Users:      userRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: systemRole}},
		Namespaces: namespaceRepoStub{},
		Roles:      roleRepoStub{permissions: granted},
		Log:        slog.New(slog.DiscardHandler),
	}
}

func tenantContext() context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, "namespace-1")
}

// TestRequireAllowsSystemAdminReadOnly 确认平台管理员不带成员身份也能通过全部租户 :read 权限。
func TestRequireAllowsSystemAdminReadOnly(t *testing.T) {
	ctx := tenantContext()
	for _, systemRole := range []string{perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin} {
		checker := checkerFor(systemRole)
		for _, code := range perm.ReadOnlyNames() {
			if err := checker.Require(ctx, code); err != nil {
				t.Fatalf("%s require %s: %v", systemRole, code, err)
			}
			ns, err := checker.Allow(ctx, code)
			if err != nil || ns != "namespace-1" {
				t.Fatalf("%s allow %s = %q, %v", systemRole, code, ns, err)
			}
		}
	}
}

// TestRequireRejectsSystemAdminWrites 确认只读放行不覆盖写、删除、介入和执行动作。
func TestRequireRejectsSystemAdminWrites(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleSystemAdmin)
	for _, code := range []string{
		perm.MemberWrite, perm.MemberDelete, perm.AlertRuleWrite,
		perm.AlertIntervene, perm.SSHExec, perm.NamespaceDelete, perm.NamespaceTransfer,
	} {
		err := checker.Require(ctx, code)
		if reason := kerrors.Reason(err); reason != merr.ReasonForbidden {
			t.Fatalf("require %s reason = %q, want %q (err = %v)", code, reason, merr.ReasonForbidden, err)
		}
	}
	// 平台权限只能走 RequireSystem，不能借租户上下文放行。
	if err := checker.Require(ctx, perm.SystemUserRead); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("system permission gated by tenant context: %v", err)
	}
}

func TestRequireKeepsMemberPermissions(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleUser, perm.RoleRead, perm.RoleWrite)
	for _, code := range []string{perm.RoleRead, perm.RoleWrite} {
		if err := checker.Require(ctx, code); err != nil {
			t.Fatalf("member require %s: %v", code, err)
		}
	}
	if err := checker.Require(ctx, perm.RoleDelete); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("member without role delete: %v", err)
	}
}

func TestRequireRejectsNormalUserOnForeignTenant(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleUser)
	if err := checker.Require(ctx, perm.AlertRuleRead); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("normal user read foreign tenant: %v", err)
	}
}

// TestRequireReadOnlyOnDisabledNamespace 确认只读切入也要遵守租户停用规则。
func TestRequireReadOnlyOnDisabledNamespace(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleSystemAdmin)
	checker.Namespaces = namespaceRepoStub{status: "DISABLED"}
	if err := checker.Require(ctx, perm.AlertRuleRead); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("read-only on disabled namespace: %v", err)
	}
}

func TestSystemAdmin(t *testing.T) {
	ctx := tenantContext()
	for _, systemRole := range []string{perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin} {
		if admin, err := checkerFor(systemRole).SystemAdmin(ctx); err != nil || !admin {
			t.Fatalf("%s admin = %v, %v", systemRole, admin, err)
		}
	}
	if admin, err := checkerFor(perm.SystemRoleUser).SystemAdmin(ctx); err != nil || admin {
		t.Fatalf("normal user admin = %v, %v", admin, err)
	}
	inactive := Checker{
		Users: userRepoStub{account: &bo.Account{UID: "user-1", Status: "BANNED", SystemRole: perm.SystemRoleSuperAdmin}},
		Log:   slog.New(slog.DiscardHandler),
	}
	if admin, err := inactive.SystemAdmin(ctx); err != nil || admin {
		t.Fatalf("banned super admin admin = %v, %v", admin, err)
	}
	missing := Checker{Users: userRepoStub{}, Log: slog.New(slog.DiscardHandler)}
	if admin, err := missing.SystemAdmin(ctx); err != nil || admin {
		t.Fatalf("missing user admin = %v, %v", admin, err)
	}
	withoutUser := actor.WithNamespace(context.Background(), "namespace-1")
	if _, err := checkerFor(perm.SystemRoleSystemAdmin).SystemAdmin(withoutUser); kerrors.Reason(err) != merr.ReasonUnauthenticated {
		t.Fatalf("system admin without user: %v", err)
	}
}
