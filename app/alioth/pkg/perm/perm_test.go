package perm

import "testing"

func TestRoleBoundaries(t *testing.T) {
	owner := set(RolePermissions(RoleOwner))
	admin := set(RolePermissions(RoleAdmin))
	member := set(RolePermissions(RoleMember))
	viewer := set(RolePermissions(RoleViewer))

	if _, ok := owner[NamespaceDelete]; !ok {
		t.Fatal("owner missing namespace delete")
	}
	if _, ok := owner[NamespaceTransfer]; !ok {
		t.Fatal("owner missing namespace transfer")
	}
	if _, ok := admin[NamespaceDelete]; ok {
		t.Fatal("admin has namespace delete")
	}
	if _, ok := admin[NamespaceTransfer]; ok {
		t.Fatal("admin has namespace transfer")
	}
	if _, ok := admin[NamespaceWrite]; !ok {
		t.Fatal("admin missing namespace write")
	}
	if _, ok := member[SSHExec]; !ok {
		t.Fatal("member missing exec permissions")
	}
	if _, ok := viewer[SSHExec]; ok {
		t.Fatal("viewer has exec permission")
	}
	for name := range viewer {
		if !IsReadOnly(name) {
			t.Fatalf("viewer permission %s is not read-only", name)
		}
	}
}

// TestK8sPermissionCatalog 确认 Kubernetes 资源级权限都在租户目录里，
// 并且 Name 能由 Resource 和 Action 拼回，前端分组和详情页才不会错位。
func TestK8sPermissionCatalog(t *testing.T) {
	catalog := set(Names())
	want := []string{
		K8sNodeRead, K8sNodeWrite,
		K8sWorkloadRead, K8sWorkloadWrite,
		K8sPodRead, K8sPodWrite, K8sPodLogs,
		K8sIngressRead, K8sServiceRead, K8sConfigMapRead, K8sSecretRead,
		K8sPersistentVolumeRead, K8sPersistentVolumeClaimRead,
		K8sJobRead, K8sJobWrite, K8sCronJobRead, K8sCronJobWrite,
	}
	for _, name := range want {
		if _, ok := catalog[name]; !ok {
			t.Fatalf("tenant catalog missing %s", name)
		}
		if !IsTenantPermission(name) {
			t.Fatalf("%s is not a tenant permission", name)
		}
	}
	for _, item := range Catalog() {
		if item.Name != item.Resource+":"+item.Action {
			t.Fatalf("permission %s name does not match resource/action %s/%s", item.Name, item.Resource, item.Action)
		}
	}
	if len(K8sReadPermissions()) != 12 {
		t.Fatalf("k8s read permissions = %d, want 12", len(K8sReadPermissions()))
	}
	if len(K8sWritePermissions()) != 5 {
		t.Fatalf("k8s write permissions = %d, want 5", len(K8sWritePermissions()))
	}
}

// TestK8sRoleBoundaries 确认 owner/admin 拿到全部资源读写，
// member/viewer 只拿到全部资源读和 Pod 日志，不含任何资源写。
func TestK8sRoleBoundaries(t *testing.T) {
	owner := set(RolePermissions(RoleOwner))
	admin := set(RolePermissions(RoleAdmin))
	member := set(RolePermissions(RoleMember))
	viewer := set(RolePermissions(RoleViewer))
	reads := K8sReadPermissions()
	if !IsReadOnly(K8sPodLogs) {
		t.Fatal("k8s pod logs is not read-only")
	}
	for _, code := range reads {
		if _, ok := member[code]; !ok {
			t.Fatalf("member missing k8s read permission %s", code)
		}
		if _, ok := viewer[code]; !ok {
			t.Fatalf("viewer missing k8s read permission %s", code)
		}
	}
	for _, code := range K8sWritePermissions() {
		if _, ok := owner[code]; !ok {
			t.Fatalf("owner missing k8s write permission %s", code)
		}
		if _, ok := admin[code]; !ok {
			t.Fatalf("admin missing k8s write permission %s", code)
		}
		if _, ok := member[code]; ok {
			t.Fatalf("member has k8s write permission %s", code)
		}
		if _, ok := viewer[code]; ok {
			t.Fatalf("viewer has k8s write permission %s", code)
		}
	}
}

// TestReadOnlyCatalog 确认只读判断覆盖 read 与 logs，写、删除、介入和执行都不算只读。
func TestReadOnlyCatalog(t *testing.T) {
	reads := set(ReadOnlyNames())
	for _, item := range Catalog() {
		_, inReads := reads[item.Name]
		want := item.Action == "read" || item.Action == "logs"
		if inReads != want {
			t.Fatalf("permission %s read-only = %v, want %v", item.Name, inReads, want)
		}
		if got := IsReadOnly(item.Name); got != want {
			t.Fatalf("IsReadOnly(%s) = %v, want %v", item.Name, got, want)
		}
	}
	for _, write := range []string{
		MemberWrite, MemberDelete, AlertIntervene, SSHExec, NamespaceDelete, NamespaceTransfer,
		K8sPodWrite, K8sCronJobWrite,
	} {
		if IsReadOnly(write) {
			t.Fatalf("write permission %s treated as read-only", write)
		}
	}
	if IsReadOnly("k8s:node") || IsReadOnly("") {
		t.Fatal("unknown permission treated as read-only")
	}
	if IsReadOnly(SystemNamespaceRead) || IsReadOnly(SystemUserRead) {
		t.Fatal("system permission treated as tenant read-only permission")
	}
	if len(reads) != len(readPermissions()) {
		t.Fatalf("read-only catalog size = %d, want %d", len(reads), len(readPermissions()))
	}
}

// TestSystemTenantIsolation 确认租户角色拿不到平台权限，平台角色也不会自动获得租户权限。
func TestSystemTenantIsolation(t *testing.T) {
	tenant := set(Names())
	for _, item := range SystemCatalog() {
		if _, ok := tenant[item.Name]; ok {
			t.Fatalf("tenant catalog contains system permission %s", item.Name)
		}
	}
	system := set(SystemNames())
	for _, item := range Catalog() {
		if _, ok := system[item.Name]; ok {
			t.Fatalf("system catalog contains tenant permission %s", item.Name)
		}
	}
	super := set(SystemPermissions(SystemRoleSuperAdmin))
	if _, ok := super[SystemUserRead]; !ok {
		t.Fatal("super admin missing system user read")
	}
	if _, ok := super[SystemRoleWrite]; !ok {
		t.Fatal("super admin missing system role write")
	}
	admin := set(SystemPermissions(SystemRoleSystemAdmin))
	if _, ok := admin[SystemUserRead]; !ok {
		t.Fatal("system admin missing system user read")
	}
	if _, ok := admin[SystemRoleWrite]; ok {
		t.Fatal("system admin has system role write")
	}
	if _, ok := admin[SystemNamespaceRead]; !ok {
		t.Fatal("system admin missing system namespace read")
	}
	if _, ok := super[SystemNamespaceRead]; !ok {
		t.Fatal("super admin missing system namespace read")
	}
	if len(SystemPermissions(SystemRoleUser)) != 0 {
		t.Fatal("normal user has system permissions")
	}
}

// TestTenantPermissionCatalog 确认租户权限目录不接受平台权限。
func TestTenantPermissionCatalog(t *testing.T) {
	if IsTenantPermission(SystemUserRead) {
		t.Fatal("system permission accepted as tenant permission")
	}
	if IsTenantPermission("user:read") {
		t.Fatal("legacy user permission accepted as tenant permission")
	}
	if !IsTenantPermission(MemberRead) {
		t.Fatal("member read is not a tenant permission")
	}
}

func TestSystemRoles(t *testing.T) {
	for _, role := range SystemRoles() {
		if !IsSystemRole(role) {
			t.Fatalf("role %s is not recognized", role)
		}
	}
	if IsSystemAdmin(SystemRoleUser) {
		t.Fatal("normal user treated as system admin")
	}
	if !IsSystemAdmin(SystemRoleSuperAdmin) || !IsSystemAdmin(SystemRoleSystemAdmin) {
		t.Fatal("system admin roles not recognized")
	}
}

func set(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}
