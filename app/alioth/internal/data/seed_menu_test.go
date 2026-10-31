package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackage"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/featurepackageitem"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

// TestSeedMigratesLegacyMenuPermissions 验证拆分前按能力生成的菜单授权会迁移到新菜单：
// 数据源菜单合并到 menu:datasource，并补默认读权限，旧菜单权限随后被清理。
func TestSeedMigratesLegacyMenuPermissions(t *testing.T) {
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

	stamp := time.Now().Format("150405.000000000")
	pkg, err := features.CreatePackage(ctx, bo.FeaturePackage{
		Name: "it-legacy-menu-pkg-" + stamp, DurationDays: 30,
		Visibility: perm.PackageVisibilityPublic, MaxMembers: 5,
		Features: []string{perm.CapabilityDatasourcePrometheus},
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	ns, err := namespaces.CreateNamespace(ctx, "it-legacy-menu-ns-"+stamp, "it", "it-legacy-owner-"+stamp, pkg.UID)
	if err != nil {
		t.Fatalf("create namespace: %v", err)
	}
	roleRow, err := store.client.Role.Create().
		SetUID("it-legacy-role-" + stamp).
		SetNamespaceUID(ns.UID).
		SetName("legacy").
		Save(ctx)
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	oldName := perm.MenuPermission(perm.CapabilityDatasourcePrometheus)
	oldPermission, err := store.client.Permission.Create().
		SetUID(id.New()).
		SetName(oldName).
		SetResource("menu").
		SetAction(perm.CapabilityDatasourcePrometheus).
		Save(ctx)
	if err != nil {
		t.Fatalf("create old permission: %v", err)
	}
	if _, err := store.client.RolePermission.Create().
		SetUID(id.New()).
		SetNamespaceUID(ns.UID).
		SetRoleUID(roleRow.UID).
		SetPermissionUID(oldPermission.UID).
		Save(ctx); err != nil {
		t.Fatalf("link old permission: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.RoleUIDEQ(roleRow.UID)).Exec(ctx)
		_, _ = store.client.Role.Delete().Where(role.UIDEQ(roleRow.UID)).Exec(ctx)
		_, _ = store.client.Permission.Delete().Where(permission.NameEQ(oldName)).Exec(ctx)
		_, _ = store.client.Namespace.Delete().Where(namespace.UIDEQ(ns.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackageItem.Delete().Where(featurepackageitem.PackageUIDEQ(pkg.UID)).Exec(ctx)
		_, _ = store.client.FeaturePackage.Delete().Where(featurepackage.UIDEQ(pkg.UID)).Exec(ctx)
	})

	// seed 可重复执行；这次会跑菜单权限迁移。
	if err := seed(ctx, store.client, logger); err != nil {
		t.Fatalf("seed: %v", err)
	}
	rows, err := store.client.RolePermission.Query().
		Where(rolepermission.RoleUIDEQ(roleRow.UID)).
		All(ctx)
	if err != nil {
		t.Fatalf("role permissions: %v", err)
	}
	permissions, err := store.client.Permission.Query().All(ctx)
	if err != nil {
		t.Fatalf("permissions: %v", err)
	}
	uidToName := make(map[string]string, len(permissions))
	for _, item := range permissions {
		uidToName[item.UID] = item.Name
	}
	held := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		held[uidToName[row.PermissionUID]] = struct{}{}
	}
	if _, ok := held[perm.MenuPermission("datasource")]; !ok {
		t.Fatalf("legacy menu not migrated: %v", held)
	}
	if _, ok := held[perm.DatasourceRead]; !ok {
		t.Fatalf("default read not added: %v", held)
	}
	if _, ok := held[oldName]; ok {
		t.Fatalf("legacy menu permission still granted: %v", held)
	}
	if exists, err := store.client.Permission.Query().Where(permission.NameEQ(oldName)).Exist(ctx); err != nil {
		t.Fatalf("old permission query: %v", err)
	} else if exists {
		t.Fatal("stale legacy menu permission was not pruned")
	}
}
