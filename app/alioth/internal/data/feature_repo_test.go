package data

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespacefeature"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// TestFeatureRepoCapabilitiesAndLifecycle 用真库覆盖能力化套餐的核心语义：
// 基础能力、只增不减、额外能力、到期与停用失效、停用原因、归档级联软删除和菜单权限补齐。
// 需要独立测试库：ALIOTH_TEST_DSN=postgres://alioth:alioth@127.0.0.1:5432/alioth_test?sslmode=disable
func TestFeatureRepoCapabilitiesAndLifecycle(t *testing.T) {
	source := testDSN(t)
	ctx := context.Background()
	if err := Migrate(ctx, source); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Fatalf("dependencies: %v", err)
	}
	t.Cleanup(cleanup)
	logger := slog.Default()
	features := NewFeatureRepo(store, logger)
	namespaces := NewNamespaceRepo(store, logger)

	base, err := features.GetPackage(ctx, basePackageUID(ctx, t, store))
	if err != nil {
		t.Fatalf("base package: %v", err)
	}
	stamp := time.Now().Format("150405.000000000")
	namespaceUID := "it-caps-ns-" + stamp
	ownerUID := "it-caps-owner-" + stamp
	t.Cleanup(func() {
		// 真库测试自己清理，避免影响下一次运行和人工排查。
		_, _ = store.client.NamespaceFeature.Delete().Where(namespacefeature.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.MemberRole.Delete().Where(memberrole.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Member.Delete().Where(member.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.NamespaceUIDEQ(namespaceUID)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(namespaceUID)).Exec(ctx)
	})

	// 套餐：30 天，能力为实时告警 + 集群接入。
	packageName := "it-caps-pkg-" + stamp
	created, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: packageName, Remark: "it", DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 5, PriceCents: 9900,
		Features: []string{perm.CapabilityAlertRealtime, perm.CapabilityClusterAccess},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(created.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(created.UID)).Exec(ctx)
	})
	if created.DurationDays != 30 {
		t.Fatalf("duration days = %d, want 30", created.DurationDays)
	}
	if len(created.Features) != 2 {
		t.Fatalf("package features = %v, want 2 capabilities", created.Features)
	}

	// 新租户绑定显式选择的套餐：先按基础套餐验证默认能力与无到期时间。
	ns, err := namespaces.CreateNamespace(ctx, namespaceUID, "it", ownerUID, base.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	state, err := features.NamespaceFeatureState(ctx, ns.UID)
	if err != nil {
		t.Fatalf("namespace feature state: %v", err)
	}
	if !state.ExpiresAt.IsZero() || state.Expired {
		t.Fatalf("new tenant expires at = %v, expired = %v", state.ExpiresAt, state.Expired)
	}
	if state.PackageName != base.Name {
		t.Fatalf("package name = %q, want %q", state.PackageName, base.Name)
	}
	capabilities, err := features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil {
		t.Fatalf("effective capabilities: %v", err)
	}
	if !sameStrings(capabilities, perm.BaseCapabilities()) {
		t.Fatalf("new tenant capabilities = %v, want base capabilities", capabilities)
	}

	// 分配套餐并额外加开 Pod 能力：目标 = 套餐能力 ∪ 额外能力，只存额外能力。
	target := []string{perm.CapabilityAlertRealtime, perm.CapabilityClusterAccess, perm.CapabilityK8sPod}
	expiresAt := time.Now().Add(30 * 24 * time.Hour).Truncate(time.Second)
	if err := features.SetNamespaceFeatureState(ctx, ns.UID, created.UID, target, expiresAt); err != nil {
		t.Fatalf("set namespace features: %v", err)
	}
	state, err = features.NamespaceFeatureState(ctx, ns.UID)
	if err != nil {
		t.Fatalf("namespace feature state: %v", err)
	}
	if !sameStrings(state.PackageFeatures, []string{perm.CapabilityAlertRealtime, perm.CapabilityClusterAccess}) {
		t.Fatalf("package features = %v", state.PackageFeatures)
	}
	if !sameStrings(state.ExtraFeatures, []string{perm.CapabilityK8sPod}) {
		t.Fatalf("extra features = %v, want only k8s_pod", state.ExtraFeatures)
	}
	if !state.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expires at = %v, want %v", state.ExpiresAt, expiresAt)
	}
	capabilities, err = features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil || !sameStrings(capabilities, target) {
		t.Fatalf("effective capabilities = %v, err = %v, want %v", capabilities, err, target)
	}
	modules, err := features.EffectiveModules(ctx, ns.UID)
	if err != nil {
		t.Fatalf("effective modules: %v", err)
	}
	if !sameStrings(modules, []string{perm.FeatureAlert, perm.FeatureCluster}) {
		t.Fatalf("effective modules = %v", modules)
	}

	// 只增不减：目标缺少套餐能力时拒绝。
	err = features.SetNamespaceFeatureState(ctx, ns.UID, created.UID, []string{perm.CapabilityAlertRealtime}, expiresAt)
	if err == nil || !strings.Contains(err.Error(), "features must include the package features") {
		t.Fatalf("missing package capability err = %v", err)
	}
	if err := features.SetNamespaceFeatureState(ctx, ns.UID, created.UID, []string{"not-a-capability"}, expiresAt); err == nil {
		t.Fatal("invalid capability accepted")
	}

	// 预置系统角色补齐了菜单权限和额外能力权限。
	ownerRole, err := store.client.Role.Query().
		Where(role.NamespaceUIDEQ(ns.UID), role.NameEQ(perm.RoleOwner), role.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		t.Fatalf("owner role: %v", err)
	}
	held := rolePermissionNames(ctx, t, store, ownerRole.UID)
	for _, code := range []string{
		perm.MenuPermission(perm.CapabilityAlertRealtime),
		perm.MenuPermission(perm.CapabilityClusterAccess),
		perm.MenuPermission(perm.CapabilityK8sPod),
		perm.K8sPodRead,
		perm.ClusterRead,
	} {
		if _, ok := held[code]; !ok {
			t.Fatalf("owner role missing %s after capability update", code)
		}
	}

	// 到期：能力生效为空，配置仍然保留并标记 Expired。
	past := time.Now().Add(-time.Hour)
	if err := features.SetNamespaceFeatureState(ctx, ns.UID, created.UID, target, past); err != nil {
		t.Fatalf("set expired features: %v", err)
	}
	capabilities, err = features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil || len(capabilities) != 0 {
		t.Fatalf("expired capabilities = %v, err = %v", capabilities, err)
	}
	state, err = features.NamespaceFeatureState(ctx, ns.UID)
	if err != nil {
		t.Fatalf("namespace feature state: %v", err)
	}
	if !state.Expired || !sameStrings(state.Features, target) {
		t.Fatalf("expired state = %+v", state)
	}

	// 停用：要求原因、写 disable_reason、有效能力为空、租户列表不再出现。
	updated, err := namespaces.SetNamespaceStatus(ctx, ns.UID, "DISABLED", "套餐到期未续费")
	if err != nil {
		t.Fatalf("disable namespace: %v", err)
	}
	if updated.Status != "DISABLED" || updated.DisableReason != "套餐到期未续费" {
		t.Fatalf("disabled namespace = %+v", updated)
	}
	capabilities, err = features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil || len(capabilities) != 0 {
		t.Fatalf("disabled capabilities = %v, err = %v", capabilities, err)
	}
	if _, err := namespaces.SetNamespaceStatus(ctx, ns.UID, "DISABLED", "  "); err == nil {
		t.Fatal("disable without reason accepted")
	}
	mine, err := namespaces.ListUserNamespaces(ctx, ownerUID)
	if err != nil {
		t.Fatalf("list user namespaces: %v", err)
	}
	if len(mine) != 0 {
		t.Fatalf("disabled namespace listed for member: %v", mine)
	}
	all, _, err := namespaces.ListAllNamespaces(ctx, 1, 100, "", "DISABLED")
	if err != nil {
		t.Fatalf("list all namespaces: %v", err)
	}
	found := false
	for _, item := range all {
		if item.UID != ns.UID {
			continue
		}
		found = true
		if item.PackageName != created.Name || item.DisableReason != "套餐到期未续费" || item.PackageExpiresAt.IsZero() {
			t.Fatalf("platform namespace row = %+v", item)
		}
	}
	if !found {
		t.Fatal("disabled namespace missing from platform list")
	}

	// 启用清空原因；把到期时间改回未来后能力恢复。
	updated, err = namespaces.SetNamespaceStatus(ctx, ns.UID, "ACTIVE", "ignored")
	if err != nil {
		t.Fatalf("enable namespace: %v", err)
	}
	if updated.Status != "ACTIVE" || updated.DisableReason != "" {
		t.Fatalf("enabled namespace = %+v", updated)
	}
	if err := features.SetNamespaceFeatureState(ctx, ns.UID, created.UID, target, expiresAt); err != nil {
		t.Fatalf("restore features: %v", err)
	}
	capabilities, err = features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil || !sameStrings(capabilities, target) {
		t.Fatalf("restored capabilities = %v, err = %v", capabilities, err)
	}

	// 归档：租户与其业务数据一起软删除，有效能力为空。
	if err := features.ArchiveNamespace(ctx, ns.UID); err != nil {
		t.Fatalf("archive namespace: %v", err)
	}
	if _, err := namespaces.GetNamespace(ctx, ns.UID); err == nil {
		t.Fatal("archived namespace still readable")
	}
	roleRows, err := store.client.Role.Query().Where(role.NamespaceUIDEQ(ns.UID)).All(ctx)
	if err != nil {
		t.Fatalf("query archived roles: %v", err)
	}
	for _, item := range roleRows {
		if item.DeletedAt == nil {
			t.Fatalf("archived namespace left role %s not deleted", item.UID)
		}
	}
	capabilities, err = features.EffectiveCapabilities(ctx, ns.UID)
	if err != nil || len(capabilities) != 0 {
		t.Fatalf("archived capabilities = %v, err = %v", capabilities, err)
	}
	if err := features.ArchiveNamespace(ctx, ns.UID); err == nil {
		t.Fatal("archived namespace archived twice without error")
	}
}

func basePackageUID(ctx context.Context, t *testing.T, store *Dependencies) string {
	t.Helper()
	row, err := store.client.FeaturePackage.Query().Where(featurepackage.IsBase(true), featurepackage.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		t.Fatalf("base package: %v", err)
	}
	return row.UID
}

func rolePermissionNames(ctx context.Context, t *testing.T, store *Dependencies, roleUID string) map[string]struct{} {
	t.Helper()
	links, err := store.client.RolePermission.Query().Where(rolepermission.RoleUIDEQ(roleUID)).All(ctx)
	if err != nil {
		t.Fatalf("role permissions: %v", err)
	}
	rows, err := store.client.Permission.Query().All(ctx)
	if err != nil {
		t.Fatalf("permissions: %v", err)
	}
	names := make(map[string]string, len(rows))
	for _, item := range rows {
		names[item.UID] = item.Name
	}
	out := make(map[string]struct{}, len(links))
	for _, link := range links {
		out[names[link.PermissionUID]] = struct{}{}
	}
	return out
}

// sameStrings 比较两个能力集合，忽略顺序，但要求元素唯一且完全一致。
func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	index := make(map[string]int, len(want))
	for _, item := range want {
		index[item]++
	}
	for _, item := range got {
		index[item]--
		if index[item] < 0 {
			return false
		}
	}
	for _, count := range index {
		if count != 0 {
			return false
		}
	}
	return true
}
