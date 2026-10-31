package role

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

type roleRepoStub struct {
	repository.RoleRepo
	permissions map[string]struct{}
	member      bool
	current     *bo.Role
	updated     []string
}

func (s *roleRepoStub) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return s.permissions, nil
}

func (s *roleRepoStub) IsMember(context.Context, string, string) (bool, error) {
	return s.member, nil
}

func (s *roleRepoStub) GetRole(context.Context, string, string) (*bo.Role, error) {
	if s.current == nil {
		return nil, merr.NotFound("not found")
	}
	return s.current, nil
}

func (s *roleRepoStub) UpdateRolePermissions(_ context.Context, _, _ string, names []string) (*bo.Role, error) {
	s.updated = names
	return s.current, nil
}

type userRepoStub struct {
	repository.UserRepo
	account *bo.Account
}

func (s userRepoStub) UserByUID(context.Context, string) (*bo.Account, error) {
	return s.account, nil
}

type namespaceRepoStub struct {
	repository.NamespaceRepo
	status    string
	ownerUID  string
	expiresAt time.Time
}

func (s namespaceRepoStub) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	status := s.status
	if status == "" {
		status = "ACTIVE"
	}
	return &bo.Namespace{UID: uid, Status: status, OwnerUID: s.ownerUID, PackageExpiresAt: s.expiresAt}, nil
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

func allCapabilities() []string {
	items := perm.Capabilities()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
}

func allModules() []string {
	items := perm.Features()
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.Code)
	}
	return out
}

func authorityFor(account *bo.Account, member bool, permissions ...string) *Authority {
	granted := make(map[string]struct{}, len(permissions))
	for _, name := range permissions {
		granted[name] = struct{}{}
	}
	return &Authority{
		roles:    &roleRepoStub{permissions: granted, member: member},
		features: featureRepoStub{capabilities: allCapabilities(), modules: allModules()},
		access: access.Checker{
			Users:      userRepoStub{account: account},
			Namespaces: namespaceRepoStub{},
			Log:        slog.New(slog.DiscardHandler),
		},
	}
}

func requestContext(namespace string) context.Context {
	ctx := actor.WithUser(context.Background(), "user-1")
	return actor.WithNamespace(ctx, namespace)
}

// TestMineReadOnlyForSystemAdmin 确认非成员的平台只读角色拿到只读目录，并被标记 readOnly。
func TestMineReadOnlyForSystemAdmin(t *testing.T) {
	for _, systemRole := range []string{
		perm.SystemRoleSuperAdmin, perm.SystemRoleSystemAdmin, perm.SystemRoleReadOnly,
	} {
		authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: systemRole}, false)
		names, capabilities, readOnly, err := authority.Mine(requestContext("namespace-1"))
		if err != nil {
			t.Fatalf("%s mine: %v", systemRole, err)
		}
		if !readOnly {
			t.Fatalf("%s without membership is not read-only", systemRole)
		}
		if !slices.Equal(names, perm.ReadOnlyNames()) {
			t.Fatalf("%s read-only names = %v, want %v", systemRole, names, perm.ReadOnlyNames())
		}
		if !slices.Equal(capabilities, allCapabilities()) {
			t.Fatalf("%s capabilities = %v, want all", systemRole, capabilities)
		}
	}
}

// TestMineMemberKeepsRolePermissions 确认成员身份优先：平台角色同时是成员时按角色权限返回。
func TestMineMemberKeepsRolePermissions(t *testing.T) {
	authority := authorityFor(
		&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin},
		true,
		perm.RoleRead, perm.MenuPermission(perm.CapabilityRole),
	)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly {
		t.Fatal("member treated as read-only")
	}
	want := []string{perm.RoleRead, perm.MenuPermission(perm.CapabilityRole)}
	slices.Sort(names)
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("member names = %v, want %v", names, want)
	}
}

// TestMineFiltersDisabledCapabilities 确认未开通能力的权限和菜单都不返回，菜单与接口范围一致。
func TestMineFiltersDisabledCapabilities(t *testing.T) {
	authority := authorityFor(
		&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser},
		true,
		perm.AlertRuleRead, perm.ClusterRead,
		perm.MenuPermission(perm.CapabilityAlertRule), perm.MenuPermission(perm.CapabilityClusterAccess),
	)
	base := featureRepoStub{capabilities: perm.BaseCapabilities(), modules: perm.ModulesOfCapabilities(perm.BaseCapabilities())}
	authority.features = base
	names, capabilities, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly {
		t.Fatal("member treated as read-only")
	}
	want := []string{perm.AlertRuleRead, perm.MenuPermission(perm.CapabilityAlertRule)}
	slices.Sort(names)
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
	if !slices.Equal(capabilities, perm.BaseCapabilities()) {
		t.Fatalf("capabilities = %v, want base", capabilities)
	}
}

func TestMineEmptyForNormalUser(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}, false)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || len(names) != 0 {
		t.Fatalf("normal user names = %v, readOnly = %v", names, readOnly)
	}
}

// TestMineMemberWithoutPermissionsIsNotReadOnly 确认「成员但没有有效权限」不会被当成只读切入。
func TestMineMemberWithoutPermissionsIsNotReadOnly(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSystemAdmin}, true)
	names, _, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || len(names) != 0 {
		t.Fatalf("member without permissions names = %v, readOnly = %v", names, readOnly)
	}
}

// TestMineDisabledNamespace 确认停用租户：成员看到空权限，平台管理员只读切入被拒绝。
func TestMineDisabledNamespace(t *testing.T) {
	disabled := namespaceRepoStub{status: "DISABLED"}
	member := authorityFor(
		&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser},
		true, perm.RoleRead, perm.MenuPermission(perm.CapabilityRole),
	)
	member.access.Namespaces = disabled
	names, capabilities, readOnly, err := member.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("member on disabled namespace: %v", err)
	}
	if readOnly || len(names) != 0 || len(capabilities) != 0 {
		t.Fatalf("disabled namespace names = %v, caps = %v, readOnly = %v", names, capabilities, readOnly)
	}
	admin := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleReadOnly}, false)
	admin.access.Namespaces = disabled
	if _, _, _, err := admin.Mine(requestContext("namespace-1")); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("admin read-only on disabled namespace: %v", err)
	}
}

// TestMineExpiredNamespace 确认套餐到期后成员和只读角色都拿不到权限，且不会误标只读。
func TestMineExpiredNamespace(t *testing.T) {
	expiredns := namespaceRepoStub{expiresAt: time.Now().Add(-time.Hour)}
	member := authorityFor(
		&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser},
		true, perm.RoleRead, perm.MenuPermission(perm.CapabilityRole),
	)
	member.access.Namespaces = expiredns
	names, capabilities, readOnly, err := member.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("mine expired namespace: %v", err)
	}
	if readOnly || len(names) != 0 || len(capabilities) != 0 {
		t.Fatalf("expired names = %v, caps = %v, readOnly = %v", names, capabilities, readOnly)
	}
	admin := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleReadOnly}, false)
	admin.access.Namespaces = expiredns
	adminNames, _, adminReadOnly, err := admin.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("admin mine expired namespace: %v", err)
	}
	if adminReadOnly || len(adminNames) != 0 {
		t.Fatalf("admin expired names = %v, readOnly = %v", adminNames, adminReadOnly)
	}
}

// TestMineWithoutNamespace 确认没有租户头时返回空列表而不是报错。
func TestMineWithoutNamespace(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleSuperAdmin}, true)
	ctx := actor.WithUser(context.Background(), "user-1")
	names, _, readOnly, err := authority.Mine(ctx)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if readOnly || names == nil || len(names) != 0 {
		t.Fatalf("names without namespace = %v, readOnly = %v", names, readOnly)
	}
}

// TestNormalizeRoleGrant 确认菜单权限默认带读权限，子权限归到已开通的菜单下。
func TestNormalizeRoleGrant(t *testing.T) {
	capabilities := []string{perm.CapabilityDatasourcePrometheus}
	modules := []string{perm.FeatureDatasource}
	got := normalizeRoleGrant(
		[]string{perm.MenuPermission("datasource")},
		capabilities,
		modules,
	)
	if !slices.Contains(got, perm.DatasourceRead) {
		t.Fatalf("menu grant missing default read: %v", got)
	}
	if slices.Contains(got, perm.DatasourceWrite) {
		t.Fatalf("menu grant added write without asking: %v", got)
	}
	// 只提交写权限时补菜单权限和默认读权限，避免出现没有菜单的子权限。
	got = normalizeRoleGrant([]string{perm.DatasourceWrite}, capabilities, modules)
	if !slices.Contains(got, perm.MenuPermission("datasource")) {
		t.Fatalf("child grant missing menu: %v", got)
	}
	if !slices.Contains(got, perm.DatasourceRead) {
		t.Fatalf("child grant missing default read: %v", got)
	}
}

// TestMineOwnerHasImplicitPermissions 确认负责人不依赖任何角色授权：
// /self/permissions 返回全部租户权限（含负责人专属的删除/移交与菜单权限），
// 但只保留套餐能力范围内的部分。
func TestMineOwnerHasImplicitPermissions(t *testing.T) {
	authority := authorityFor(&bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}, false)
	authority.access.Namespaces = namespaceRepoStub{ownerUID: "user-1"}
	authority.features = featureRepoStub{
		capabilities: perm.BaseCapabilities(),
		modules:      perm.ModulesOfCapabilities(perm.BaseCapabilities()),
	}
	names, capabilities, readOnly, err := authority.Mine(requestContext("namespace-1"))
	if err != nil {
		t.Fatalf("owner mine: %v", err)
	}
	if readOnly {
		t.Fatal("owner treated as read-only")
	}
	for _, code := range []string{
		perm.NamespaceDelete, perm.NamespaceTransfer, perm.AlertRuleWrite,
		perm.MemberWrite, perm.MenuPermission(perm.CapabilityAlertRule),
		perm.MenuPermission("namespace"),
	} {
		if !slices.Contains(names, code) {
			t.Fatalf("owner permissions missing %s: %v", code, names)
		}
	}
	if slices.Contains(names, perm.MenuPermission(perm.CapabilityClusterAccess)) {
		t.Fatal("owner got menu permission outside licensed capabilities")
	}
	if !slices.Equal(capabilities, perm.BaseCapabilities()) {
		t.Fatalf("owner capabilities = %v, want base", capabilities)
	}
}

// authorityForCaller 构造一个角色授权用例：namespace 负责人和调用者权限可分别设置。
func authorityForCaller(ownerUID string, current *bo.Role, held ...string) (*Authority, *roleRepoStub) {
	granted := make(map[string]struct{}, len(held))
	for _, name := range held {
		granted[name] = struct{}{}
	}
	stub := &roleRepoStub{permissions: granted, member: true, current: current}
	authority := &Authority{
		roles:    stub,
		features: featureRepoStub{capabilities: allCapabilities(), modules: allModules()},
		access: access.Checker{
			Users:      userRepoStub{account: &bo.Account{UID: "user-1", Status: "ACTIVE", SystemRole: perm.SystemRoleUser}},
			Namespaces: namespaceRepoStub{ownerUID: ownerUID},
			Roles:      stub,
			Features:   featureRepoStub{capabilities: allCapabilities(), modules: allModules()},
			Log:        slog.New(slog.DiscardHandler),
		},
	}
	return authority, stub
}

// TestUpdateRolePermissionsRejectsBeyondOwn 确认非负责人只能授予自己持有的权限。
func TestUpdateRolePermissionsRejectsBeyondOwn(t *testing.T) {
	current := &bo.Role{UID: "role-1", Name: "ops"}
	authority, _ := authorityForCaller("", current,
		perm.RoleRead, perm.RoleWrite, perm.MenuPermission(perm.CapabilityRole),
	)
	_, err := authority.UpdateRolePermissions(requestContext("namespace-1"), "role-1", []string{perm.MemberRead})
	if err == nil || !strings.Contains(kerrors.FromError(err).Message, "cannot grant permissions beyond your own") {
		t.Fatalf("grant beyond own err = %v, want subset rejection", err)
	}
}

// TestUpdateRolePermissionsRejectsNormalizeAddedBeyondOwn 确认 normalizeRoleGrant
// 自动补出的默认读权限也要在调用者权限内，不能借补全绕过子集校验。
func TestUpdateRolePermissionsRejectsNormalizeAddedBeyondOwn(t *testing.T) {
	current := &bo.Role{UID: "role-1", Name: "ops"}
	authority, _ := authorityForCaller("", current,
		perm.RoleRead, perm.RoleWrite, perm.MenuPermission(perm.CapabilityRole),
		perm.DatasourceWrite, perm.MenuPermission("datasource"),
	)
	_, err := authority.UpdateRolePermissions(requestContext("namespace-1"), "role-1", []string{perm.DatasourceWrite})
	if err == nil || !strings.Contains(kerrors.FromError(err).Message, "cannot grant permissions beyond your own") {
		t.Fatalf("grant with normalize-added read err = %v, want subset rejection", err)
	}
}

// TestUpdateRolePermissionsOwnerBypassesSubset 确认负责人不受子集限制，
// 但仍受租户已开通能力范围限制。
func TestUpdateRolePermissionsOwnerBypassesSubset(t *testing.T) {
	current := &bo.Role{UID: "role-1", Name: "ops"}
	authority, stub := authorityForCaller("user-1", current)
	if _, err := authority.UpdateRolePermissions(requestContext("namespace-1"), "role-1", []string{perm.MemberWrite}); err != nil {
		t.Fatalf("owner grant: %v", err)
	}
	if !slices.Contains(stub.updated, perm.MemberWrite) {
		t.Fatalf("owner grant not written: %v", stub.updated)
	}
	authority.features = featureRepoStub{
		capabilities: perm.BaseCapabilities(),
		modules:      perm.ModulesOfCapabilities(perm.BaseCapabilities()),
	}
	_, err := authority.UpdateRolePermissions(requestContext("namespace-1"), "role-1", []string{perm.ClusterWrite})
	if err == nil || kerrors.FromError(err).Message != "feature not enabled" {
		t.Fatalf("owner grant out-of-package err = %v, want feature not enabled", err)
	}
}
