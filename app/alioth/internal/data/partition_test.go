package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/permission"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/role"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/rolepermission"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
)

func TestWeekStart(t *testing.T) {
	from, until := weekStart(time.Date(2026, 10, 21, 13, 45, 0, 0, time.UTC))
	if from.Weekday() != time.Monday || from.Hour() != 0 || from.Minute() != 0 {
		t.Fatalf("week start = %s", from)
	}
	if until.Sub(from) != 7*24*time.Hour {
		t.Fatalf("week span = %s", until.Sub(from))
	}
	if !from.Before(until) || !until.After(from) {
		t.Fatalf("invalid week range %s - %s", from, until)
	}
}

func TestPartitionIndexesCovered(t *testing.T) {
	for _, table := range partitionedTables {
		if len(partitionIndexes[table]) == 0 {
			t.Fatalf("%s has no partition index statements", table)
		}
	}
}

// TestDropStalePermissionsExpandsLegacyK8sPermissions 确认 M25 升级时旧角色不会丢权限：
// ApplySchema 先补齐权限目录，再把 k8s:*:write 授权展开成操作权限后才删除旧权限行。
// 新操作 Pod Exec 不在展开表里，不会被旧写权限带出来；重复执行保持结果不变。
func TestDropStalePermissionsExpandsLegacyK8sPermissions(t *testing.T) {
	source := testDSN(t)
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	ctx := context.Background()
	if err := ApplySchema(ctx, store.client, store.db); err != nil {
		t.Fatal(err)
	}

	suffix := time.Now().Format("150405.000000")
	roleUID := "m25-legacy-role-" + suffix
	namespaceUID := "m25-legacy-ns-" + suffix
	legacyPermissions := []struct {
		uid, name, resource, action string
	}{
		{"m25-legacy-workload-" + suffix, "k8s:workload:write", "k8s:workload", "write"},
		{"m25-legacy-node-" + suffix, "k8s:node:write", "k8s:node", "write"},
	}
	for _, item := range legacyPermissions {
		if _, err := store.client.Permission.Create().
			SetUID(item.uid).
			SetName(item.name).
			SetResource(item.resource).
			SetAction(item.action).
			Save(ctx); err != nil {
			t.Fatalf("create legacy permission %s: %v", item.name, err)
		}
	}
	roleRow, err := store.client.Role.Create().
		SetUID(roleUID).
		SetNamespaceUID(namespaceUID).
		SetName(roleUID).
		SetIsSystem(false).
		Save(ctx)
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	t.Cleanup(func() {
		_, _ = store.client.RolePermission.Delete().Where(rolepermission.RoleUIDEQ(roleUID)).Exec(context.Background())
		_, _ = store.client.Role.Delete().Where(role.UIDEQ(roleRow.UID)).Exec(context.Background())
	})
	for index, item := range legacyPermissions {
		if _, err := store.client.RolePermission.Create().
			SetUID("m25-legacy-rp-" + suffix + "-" + item.resource).
			SetNamespaceUID(namespaceUID).
			SetRoleUID(roleUID).
			SetPermissionUID(item.uid).
			Save(ctx); err != nil {
			t.Fatalf("link legacy permission %d: %v", index, err)
		}
	}

	assertRolePermissions := func() {
		t.Helper()
		links, err := store.client.RolePermission.Query().Where(rolepermission.RoleUIDEQ(roleUID)).All(ctx)
		if err != nil {
			t.Fatalf("query role permissions: %v", err)
		}
		uids := make([]string, 0, len(links))
		for _, link := range links {
			uids = append(uids, link.PermissionUID)
		}
		rows, err := store.client.Permission.Query().Where(permission.UIDIn(uids...)).All(ctx)
		if err != nil {
			t.Fatalf("query permissions: %v", err)
		}
		got := make(map[string]struct{}, len(rows))
		for _, row := range rows {
			got[row.Name] = struct{}{}
		}
		want := []string{perm.K8sNodeCordon, perm.K8sWorkloadScale, perm.K8sWorkloadRestart, perm.K8sWorkloadImage, perm.K8sWorkloadRollback}
		if len(got) != len(want) {
			t.Fatalf("role permissions = %v, want %v", got, want)
		}
		for _, name := range want {
			if _, ok := got[name]; !ok {
				t.Fatalf("role missing expanded permission %s, got %v", name, got)
			}
		}
		if _, ok := got[perm.K8sPodExec]; ok {
			t.Fatal("pod exec must not be granted by legacy k8s:*:write")
		}
	}

	if err := dropStalePermissions(ctx, store.client); err != nil {
		t.Fatalf("expand legacy permissions: %v", err)
	}
	assertRolePermissions()
	// 幂等：再跑一次不会重复插入或删掉已展开的授权。
	if err := dropStalePermissions(ctx, store.client); err != nil {
		t.Fatalf("expand legacy permissions twice: %v", err)
	}
	assertRolePermissions()
	for _, item := range legacyPermissions {
		exists, err := store.client.Permission.Query().Where(permission.NameEQ(item.name)).Exist(ctx)
		if err != nil {
			t.Fatalf("query legacy permission %s: %v", item.name, err)
		}
		if exists {
			t.Fatalf("legacy permission %s still exists", item.name)
		}
	}
}
