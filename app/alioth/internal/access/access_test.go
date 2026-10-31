package access

import (
	"context"
	"log/slog"
	"slices"
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
	status   string
	ownerUID string
}

func (s namespaceRepoStub) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	status := s.status
	if status == "" {
		status = "ACTIVE"
	}
	return &bo.Namespace{UID: uid, Status: status, OwnerUID: s.ownerUID}, nil
}

type roleRepoStub struct {
	repository.RoleRepo
	permissions map[string]struct{}
}

func (s roleRepoStub) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return s.permissions, nil
}

type featureRepoStub struct {
	repository.FeatureRepo
	capabilities []string
	modules      []string
}

func (s featureRepoStub) EffectiveCapabilities(context.Context, string) ([]string, error) {
	return s.capabilities, nil
}

func (s featureRepoStub) EffectiveModules(context.Context, string) ([]string, error) {
	return s.modules, nil
}

// allCapabilities 返回全部能力编码，作为「租户开通了所有能力」的测试默认值。
func allCapabilities() []string {
	items := perm.Capabilities()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
}

// allModules 返回全部模块编码，作为「全能力开通」对应的模块集合。
func allModules() []string {
	items := perm.Features()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
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
		// 默认租户已开通全部能力，权限相关的用例只关心权限判断本身。
		Features: featureRepoStub{capabilities: allCapabilities(), modules: allModules()},
		Log:      slog.New(slog.DiscardHandler),
	}
}

func tenantContext() context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, "namespace-1")
}

// TestRequireAllowsSystemAdminReadOnly 确认平台只读角色不带成员身份也能通过全部租户只读权限。
func TestRequireAllowsSystemAdminReadOnly(t *testing.T) {
	ctx := tenantContext()
	for _, systemRole := range []string{
		perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly,
	} {
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
	for _, systemRole := range []string{
		perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly,
	} {
		checker := checkerFor(systemRole)
		for _, code := range []string{
			perm.MemberWrite, perm.MemberDelete, perm.AlertRuleWrite,
			perm.AlertRealtimeIntervene, perm.SSHExec, perm.NamespaceDelete, perm.NamespaceTransfer,
		} {
			err := checker.Require(ctx, code)
			if reason := kerrors.Reason(err); reason != merr.ReasonForbidden {
				t.Fatalf("%s require %s reason = %q, want %q (err = %v)", systemRole, code, reason, merr.ReasonForbidden, err)
			}
		}
		// 平台权限只能走 RequireSystem，不能借租户上下文放行。
		if err := checker.Require(ctx, perm.SystemUserRead); kerrors.Reason(err) != merr.ReasonForbidden {
			t.Fatalf("%s system permission gated by tenant context: %v", systemRole, err)
		}
	}
}

func TestRequireKeepsMemberPermissions(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleUser,
		perm.RoleRead, perm.RoleWrite, perm.MenuPermission(perm.CapabilityRole),
	)
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
	checker := checkerFor(perm.SystemRoleReadOnly)
	checker.Namespaces = namespaceRepoStub{status: "DISABLED"}
	if err := checker.Require(ctx, perm.AlertRuleRead); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("read-only on disabled namespace: %v", err)
	}
}

// TestRequireChecksFeature 确认权限通过后仍要求租户开通对应模块，系统管理员只读同样受限。
func TestRequireChecksFeature(t *testing.T) {
	ctx := tenantContext()
	base := featureRepoStub{capabilities: perm.BaseCapabilities(), modules: perm.ModulesOfCapabilities(perm.BaseCapabilities())}
	checker := checkerFor(perm.SystemRoleUser,
		perm.AlertRuleRead, perm.ClusterRead,
		perm.MenuPermission(perm.CapabilityAlertRule), perm.MenuPermission(perm.CapabilityClusterAccess),
	)
	checker.Features = base
	if err := checker.Require(ctx, perm.AlertRuleRead); err != nil {
		t.Fatalf("base capability permission: %v", err)
	}
	err := checker.Require(ctx, perm.ClusterRead)
	if reason := kerrors.Reason(err); reason != merr.ReasonForbidden || kerrors.FromError(err).Message != "feature not enabled" {
		t.Fatalf("disabled feature permission err = %v, want forbidden feature not enabled", err)
	}
	admin := checkerFor(perm.SystemRoleSystemAdmin)
	admin.Features = base
	if err := admin.Require(ctx, perm.ClusterRead); kerrors.FromError(err).Message != "feature not enabled" {
		t.Fatalf("system admin read-only disabled feature err = %v, want forbidden feature not enabled", err)
	}
	if err := admin.Require(ctx, perm.AlertRuleRead); err != nil {
		t.Fatalf("system admin read-only base feature: %v", err)
	}
}

// TestRequireCapability 确认类型级检查只在租户开通对应能力时放行。
func TestRequireCapability(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleUser)
	checker.Features = featureRepoStub{
		capabilities: []string{perm.CapabilityDatasourcePrometheus},
		modules:      []string{perm.FeatureDatasource},
	}
	if err := checker.RequireCapability(ctx, perm.CapabilityDatasourcePrometheus); err != nil {
		t.Fatalf("require authorized capability: %v", err)
	}
	err := checker.RequireCapability(ctx, perm.CapabilityDatasourceLoki)
	if err == nil || kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("require unauthorized capability err = %v, want forbidden", err)
	}
	capabilities, err := checker.Capabilities(ctx)
	if err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	if !slices.Equal(capabilities, []string{perm.CapabilityDatasourcePrometheus}) {
		t.Fatalf("capabilities = %v", capabilities)
	}
}

// TestRequireChecksMenuPermission 确认成员即使持有接口权限，也要持有对应能力的菜单权限。
func TestRequireChecksMenuPermission(t *testing.T) {
	ctx := tenantContext()
	checker := checkerFor(perm.SystemRoleUser, perm.RoleRead)
	err := checker.Require(ctx, perm.RoleRead)
	if reason := kerrors.Reason(err); reason != merr.ReasonForbidden || kerrors.FromError(err).Message != "menu permission denied" {
		t.Fatalf("missing menu permission err = %v, want forbidden menu permission denied", err)
	}
	granted := checkerFor(perm.SystemRoleUser, perm.RoleRead, perm.MenuPermission(perm.CapabilityRole))
	if err := granted.Require(ctx, perm.RoleRead); err != nil {
		t.Fatalf("role read with menu permission: %v", err)
	}
}

// TestRequireSkipsMenuPermissionForSystemAdminReadOnly 平台管理员跨租户只读不要求租户菜单授权。
func TestRequireSkipsMenuPermissionForSystemAdminReadOnly(t *testing.T) {
	ctx := tenantContext()
	for _, systemRole := range []string{
		perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly,
	} {
		if err := checkerFor(systemRole).Require(ctx, perm.RoleRead); err != nil {
			t.Fatalf("%s read-only require role:read: %v", systemRole, err)
		}
	}
}

// TestRequireOwnerHasImplicitPermissions 确认租户负责人不需要任何角色授权就持有
// 全部租户权限（含菜单权限），但未开通能力/模块仍然拒绝。
func TestRequireOwnerHasImplicitPermissions(t *testing.T) {
	ctx := tenantContext()
	// 负责人没有任何角色权限，角色仓储返回空集合。
	owner := checkerFor(perm.SystemRoleUser)
	owner.Namespaces = namespaceRepoStub{ownerUID: "user-1"}
	for _, code := range perm.Names() {
		if err := owner.Require(ctx, code); err != nil {
			t.Fatalf("owner require %s: %v", code, err)
		}
	}
	// 套餐未开通的模块对负责人同样返回 403。
	owner.Features = featureRepoStub{
		capabilities: perm.BaseCapabilities(),
		modules:      perm.ModulesOfCapabilities(perm.BaseCapabilities()),
	}
	err := owner.Require(ctx, perm.ClusterRead)
	if reason := kerrors.Reason(err); reason != merr.ReasonForbidden || kerrors.FromError(err).Message != "feature not enabled" {
		t.Fatalf("owner on disabled feature err = %v, want forbidden feature not enabled", err)
	}
	// 停用租户对负责人也是 403，负责人不会绕过租户状态。
	owner.Namespaces = namespaceRepoStub{status: "DISABLED", ownerUID: "user-1"}
	if err := owner.Require(ctx, perm.RoleRead); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("owner on disabled namespace: %v", err)
	}
}

// TestHeldPermissions 确认负责人隐式持有全部租户权限，非负责人只拿到角色权限。
func TestHeldPermissions(t *testing.T) {
	ctx := tenantContext()
	owner := checkerFor(perm.SystemRoleUser)
	owner.Namespaces = namespaceRepoStub{ownerUID: "user-1"}
	held, isOwner, err := owner.HeldPermissions(ctx)
	if err != nil {
		t.Fatalf("owner held permissions: %v", err)
	}
	if !isOwner || len(held) != len(perm.Names()) {
		t.Fatalf("owner held = %d permissions, isOwner = %v", len(held), isOwner)
	}
	member := checkerFor(perm.SystemRoleUser, perm.RoleRead)
	held, isOwner, err = member.HeldPermissions(ctx)
	if err != nil {
		t.Fatalf("member held permissions: %v", err)
	}
	if isOwner {
		t.Fatal("member treated as owner")
	}
	if _, ok := held[perm.RoleRead]; !ok || len(held) != 1 {
		t.Fatalf("member held = %v", held)
	}
}

func TestSystemAdmin(t *testing.T) {
	ctx := tenantContext()
	for _, systemRole := range []string{
		perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly,
	} {
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

// TestCurrentNamespace 确认 /self/permissions 能拿到租户状态；没有租户头时返回 nil。
func TestCurrentNamespace(t *testing.T) {
	checker := checkerFor(perm.SystemRoleUser)
	item, err := checker.CurrentNamespace(tenantContext())
	if err != nil || item == nil || item.Status != "ACTIVE" {
		t.Fatalf("current namespace = %v, %v", item, err)
	}
	missing, err := checker.CurrentNamespace(actor.WithUser(context.Background(), "user-1"))
	if err != nil || missing != nil {
		t.Fatalf("current namespace without header = %v, %v", missing, err)
	}
}

// TestEveryTenantPermissionBelongsToCapability 确认每条租户权限都挂在某个能力上，
// 否则 role/access 的能力检查会让这条权限永远无法通过菜单校验。
func TestEveryTenantPermissionBelongsToCapability(t *testing.T) {
	capabilities := perm.Capabilities()
	for _, item := range perm.Catalog() {
		if item.Resource == "menu" || item.Resource == "namespace" {
			continue
		}
		covered := slices.ContainsFunc(capabilities, func(capability perm.Capability) bool {
			return slices.Contains(capability.Permissions, item.Name)
		})
		if !covered {
			t.Fatalf("permission %s belongs to no capability", item.Name)
		}
	}
}

// TestReadOnlyNamesCoverCapabilities 确认只读目录包含菜单权限，平台只读能展示菜单。
func TestReadOnlyNamesCoverCapabilities(t *testing.T) {
	names := perm.ReadOnlyNames()
	for _, menu := range perm.Menus() {
		if !slices.Contains(names, perm.MenuPermission(menu.Code)) {
			t.Fatalf("read-only names missing menu permission for %s", menu.Code)
		}
	}
}
