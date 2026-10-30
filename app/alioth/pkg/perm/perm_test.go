package perm

import (
	"strings"
	"testing"
)

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
		want := item.Action == "read" || item.Action == "logs" || item.Resource == "menu"
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
	if _, ok := super[SystemNamespaceWrite]; !ok {
		t.Fatal("super admin missing system namespace write")
	}
	if _, ok := admin[SystemNamespaceWrite]; ok {
		t.Fatal("system admin has system namespace write")
	}
	readOnly := set(SystemPermissions(SystemRoleReadOnly))
	for _, item := range SystemCatalog() {
		isRead := strings.HasSuffix(item.Name, ":read")
		_, held := readOnly[item.Name]
		if held != isRead {
			t.Fatalf("system readonly role permission %s held = %v, want %v", item.Name, held, isRead)
		}
	}
	if _, ok := readOnly[SystemPackageWrite]; ok {
		t.Fatal("system readonly role has system package write")
	}
	if !CanReadAnyTenant(SystemRoleReadOnly) || !CanReadAnyTenant(SystemRoleSystemAdmin) || !CanReadAnyTenant(SystemRoleSuperAdmin) {
		t.Fatal("read-only tenant access not granted to platform roles")
	}
	if CanReadAnyTenant(SystemRoleUser) {
		t.Fatal("normal user can read any tenant")
	}
	if IsSystemAdmin(SystemRoleReadOnly) {
		t.Fatal("system readonly role treated as system admin")
	}
	if !IsSystemReadOnly(SystemRoleReadOnly) {
		t.Fatal("system readonly role not recognized")
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
	if !IsSystemRole(SystemRoleReadOnly) {
		t.Fatal("system readonly role not recognized")
	}
}

// TestCapabilityCatalog 确认两级功能模型：37 个能力都挂在 10 个模块下，
// 权限都在租户目录里，菜单权限与能力一一对应。
func TestCapabilityCatalog(t *testing.T) {
	items := Capabilities()
	if len(items) != 37 {
		t.Fatalf("capabilities = %d, want 37", len(items))
	}
	if items[0].Code != CapabilityAlertRealtime || items[len(items)-1].Code != CapabilityWebShell {
		t.Fatalf("unexpected capability order: first %s last %s", items[0].Code, items[len(items)-1].Code)
	}
	catalog := set(Names())
	modules := set(FeatureCodes())
	for _, item := range items {
		if !IsCapability(item.Code) {
			t.Fatalf("capability %s not recognized", item.Code)
		}
		if _, ok := modules[item.Module]; !ok {
			t.Fatalf("capability %s module %s is not a feature", item.Code, item.Module)
		}
		if len(item.Permissions) == 0 {
			t.Fatalf("capability %s has no permissions", item.Code)
		}
		for _, name := range item.Permissions {
			if _, ok := catalog[name]; !ok {
				t.Fatalf("capability %s permission %s is not in tenant catalog", item.Code, name)
			}
			if !IsTenantPermission(name) {
				t.Fatalf("capability permission %s is not a tenant permission", name)
			}
		}
		if got := FeatureOf(MenuPermission(item.Code)); got != item.Module {
			t.Fatalf("FeatureOf(menu:%s) = %s, want %s", item.Code, got, item.Module)
		}
		if got := FeatureOf(item.Code); got != item.Module {
			t.Fatalf("FeatureOf(%s) = %s, want %s", item.Code, got, item.Module)
		}
		if !IsMenuPermission(MenuPermission(item.Code)) {
			t.Fatalf("menu permission for %s not recognized", item.Code)
		}
		if !IsReadOnly(MenuPermission(item.Code)) {
			t.Fatalf("menu permission for %s is not read-only", item.Code)
		}
	}
	if len(set(MenuPermissions())) != len(items) {
		t.Fatal("menu permission count does not match capability count")
	}
	base := BaseCapabilities()
	if len(base) != 22 {
		t.Fatalf("base capabilities = %d, want 22", len(base))
	}
	// 数据源、探测和通知渠道按类型展开后，基础套餐要覆盖全部现有类型。
	if got := DatasourceDriversOf(base); len(got) != 6 {
		t.Fatalf("base datasource drivers = %v, want 6", got)
	}
	if got := ProbeDriversOf(base); len(got) != 1 || got[0] != "PROBE_HTTP" {
		t.Fatalf("base probe drivers = %v, want [PROBE_HTTP]", got)
	}
	if got := ChannelTypesOf(base); len(got) != 3 {
		t.Fatalf("base channel types = %v, want 3", got)
	}
	baseModules := set(BaseFeatures())
	for _, code := range base {
		item, _ := CapabilityOf(code)
		if _, ok := baseModules[item.Module]; !ok {
			t.Fatalf("base capability %s module %s is not a base feature", code, item.Module)
		}
	}
	if got := CapabilityPermissions([]string{CapabilityNotifyChannelSMTP, CapabilityNotifyGroup}); len(got) != 3 {
		t.Fatalf("notification permissions = %v, want 3 unique", got)
	}
	if got := CapabilityCodesOfModule(FeatureCluster); len(got) != 12 {
		t.Fatalf("cluster capabilities = %d, want 12", len(got))
	}
}

// TestRolePermissionsForCapabilities 确认预置角色按能力范围取权限：
// 模块权限与能力权限取交集，再带上对应菜单权限；namespace:* 始终保留。
func TestRolePermissionsForCapabilities(t *testing.T) {
	caps := []string{CapabilityAlertRealtime, CapabilityClusterAccess}
	owner := set(RolePermissionsForCapabilities(RoleOwner, caps))
	for _, name := range []string{AlertRead, AlertIntervene, MenuPermission(CapabilityAlertRealtime), MenuPermission(CapabilityClusterAccess), NamespaceWrite} {
		if _, ok := owner[name]; !ok {
			t.Fatalf("owner missing %s", name)
		}
	}
	for _, name := range []string{AlertWrite, AlertDelete, K8sNodeRead, MenuPermission(CapabilityAlertHistory)} {
		if _, ok := owner[name]; ok {
			t.Fatalf("owner has out-of-scope permission %s", name)
		}
	}
	if _, ok := owner[ClusterWrite]; !ok {
		t.Fatal("owner missing cluster write from cluster access capability")
	}
	viewer := set(RolePermissionsForCapabilities(RoleViewer, caps))
	for name := range viewer {
		if !IsReadOnly(name) {
			t.Fatalf("viewer permission %s is not read-only", name)
		}
	}
	member := set(RolePermissionsForCapabilities(RoleMember, []string{CapabilityWebShell}))
	if _, ok := member[SSHExec]; !ok {
		t.Fatal("member missing ssh exec in web shell capability")
	}
	if _, ok := member[MenuPermission(CapabilityWebShell)]; !ok {
		t.Fatal("member missing web shell menu permission")
	}
	if _, ok := member[MenuPermission(CapabilitySSHCommand)]; ok {
		t.Fatal("member has menu permission outside capabilities")
	}
	if _, ok := set(RolePermissionsForCapabilities(RoleViewer, nil))[MemberRead]; ok {
		t.Fatal("capability-less role kept module permissions")
	}
}

// TestEffectiveCapabilities 确认套餐能力与额外能力求并集并按目录顺序返回，
// 目录外的旧编码被忽略。
func TestEffectiveCapabilities(t *testing.T) {
	got := EffectiveCapabilities(
		[]string{CapabilityRole, "alert", CapabilityAlertRealtime},
		[]string{CapabilityRole, CapabilityK8sPod},
	)
	want := []string{CapabilityAlertRealtime, CapabilityRole, CapabilityK8sPod}
	if len(got) != len(want) {
		t.Fatalf("effective capabilities = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("effective capabilities = %v, want %v", got, want)
		}
	}
	modules := ModulesOfCapabilities([]string{CapabilityK8sPod, CapabilityClusterAccess, CapabilityMember})
	if len(modules) != 2 || modules[0] != FeatureMember || modules[1] != FeatureCluster {
		t.Fatalf("modules = %v, want [member cluster]", modules)
	}
}

func set(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}
