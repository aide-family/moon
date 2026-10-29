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
		if len(name) < 5 || name[len(name)-5:] != ":read" {
			t.Fatalf("viewer permission %s is not read", name)
		}
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

// TestReadOnlyCatalog 确认只读放行只认租户目录里的 read 动作。
// 平台管理员跨租户只读，写、删除、介入和执行动作不在其中。
func TestReadOnlyCatalog(t *testing.T) {
	reads := set(ReadOnlyNames())
	for _, item := range Catalog() {
		_, inReads := reads[item.Name]
		want := item.Action == "read"
		if inReads != want {
			t.Fatalf("permission %s read-only = %v, want %v", item.Name, inReads, want)
		}
		if got := IsReadOnly(item.Name); got != want {
			t.Fatalf("IsReadOnly(%s) = %v, want %v", item.Name, got, want)
		}
	}
	for _, write := range []string{MemberWrite, MemberDelete, AlertIntervene, SSHExec, NamespaceDelete, NamespaceTransfer} {
		if IsReadOnly(write) {
			t.Fatalf("write permission %s treated as read-only", write)
		}
	}
	if IsReadOnly(SystemNamespaceRead) || IsReadOnly(SystemUserRead) {
		t.Fatal("system permission treated as tenant read-only permission")
	}
	if IsReadOnly("member") || IsReadOnly("") {
		t.Fatal("unknown permission treated as read-only")
	}
	if len(reads) != len(readPermissions()) {
		t.Fatalf("read-only catalog size = %d, want %d", len(reads), len(readPermissions()))
	}
}

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
