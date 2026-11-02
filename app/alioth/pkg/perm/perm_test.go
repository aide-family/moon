package perm

import (
	"slices"
	"strings"
	"testing"
)

// TestK8sPermissionCatalog 确认 Kubernetes 资源级权限都在租户目录里，
// 并且 Name 能由 Resource 和 Action 拼回，前端分组和详情页才不会错位。
func TestK8sPermissionCatalog(t *testing.T) {
	catalog := set(Names())
	want := []string{
		K8sNodeRead, K8sNodeCordon,
		K8sWorkloadRead, K8sWorkloadScale, K8sWorkloadRestart, K8sWorkloadImage, K8sWorkloadRollback,
		K8sPodRead, K8sPodLogs, K8sPodDelete, K8sPodExec,
		K8sIngressRead, K8sServiceRead, K8sConfigMapRead, K8sSecretRead,
		K8sPersistentVolumeRead, K8sPersistentVolumeClaimRead,
		K8sJobRead, K8sJobDelete, K8sCronJobRead, K8sCronJobSuspend, K8sCronJobTrigger, K8sCronJobDelete,
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
	// 旧 cluster:write 展开的操作权限共 10 个；Pod 终端是新权限，不在其中。
	if len(K8sWritePermissions()) != 10 {
		t.Fatalf("k8s write permissions = %d, want 10", len(K8sWritePermissions()))
	}
	if slices.Contains(K8sWritePermissions(), K8sPodExec) {
		t.Fatal("pod exec must not be granted by the legacy cluster:write expansion")
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
		MemberWrite, MemberDelete, AlertRealtimeIntervene, SSHExec, NamespaceDelete, NamespaceTransfer,
		K8sPodDelete, K8sPodExec, K8sCronJobDelete, K8sNodeCordon, K8sWorkloadRollback,
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

// TestCapabilityCatalog 确认两级功能模型：62 个能力都挂在 11 个模块下，
// 权限都在租户目录里，菜单权限与能力一一对应。
func TestCapabilityCatalog(t *testing.T) {
	items := Capabilities()
	if len(items) != 62 {
		t.Fatalf("capabilities = %d, want 62", len(items))
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
		if got := FeatureOf(item.Code); got != item.Module {
			t.Fatalf("FeatureOf(%s) = %s, want %s", item.Code, got, item.Module)
		}
	}
	menus := Menus()
	for _, menu := range menus {
		if !IsMenu(menu.Code) {
			t.Fatalf("menu %s not recognized", menu.Code)
		}
		if got := FeatureOf(MenuPermission(menu.Code)); got != menu.Group {
			t.Fatalf("FeatureOf(menu:%s) = %s, want %s", menu.Code, got, menu.Group)
		}
		if !IsMenuPermission(MenuPermission(menu.Code)) {
			t.Fatalf("menu permission for %s not recognized", menu.Code)
		}
		if !IsReadOnly(MenuPermission(menu.Code)) {
			t.Fatalf("menu permission for %s is not read-only", menu.Code)
		}
		for _, read := range menu.DefaultReads {
			if !slices.Contains(menu.Permissions, read) {
				t.Fatalf("menu %s default read %s is not one of its permissions", menu.Code, read)
			}
		}
	}
	if len(set(MenuPermissions())) != len(menus) {
		t.Fatal("menu permission count does not match menu count")
	}
	base := BaseCapabilities()
	if len(base) != 35 {
		t.Fatalf("base capabilities = %d, want 35", len(base))
	}
	// 数据源、探测和通知渠道按类型展开后，基础套餐要覆盖全部现有类型。
	if got := DatasourceDriversOf(base); len(got) != 6 {
		t.Fatalf("base datasource drivers = %v, want 6", got)
	}
	if got := ProbeDriversOf(base); len(got) != 5 || got[0] != "PROBE_HTTP" || got[1] != "PROBE_TCP" || got[2] != "PROBE_ICMP" || got[3] != "PROBE_DNS" || got[4] != "PROBE_TLS" {
		t.Fatalf("base probe drivers = %v, want [PROBE_HTTP PROBE_TCP PROBE_ICMP PROBE_DNS PROBE_TLS]", got)
	}
	if got := ChannelTypesOf(base); len(got) != 3 {
		t.Fatalf("base channel types = %v, want 3", got)
	}
	// Webhook 按应用、短信按提供商拆分后，每个组合对应一个能力；腾讯云短信只上架能力。
	if got := ChannelScopesOf(base); len(got) != 6 {
		t.Fatalf("base channel scopes = %v, want 6", got)
	}
	if got := ChannelCapabilityOf("WEBHOOK", "DINGTALK"); got != CapabilityNotifyChannelWebhookDingTalk {
		t.Fatalf("dingtalk webhook capability = %s", got)
	}
	if got := ChannelCapabilityOf("WEBHOOK", "OTHER"); got != CapabilityNotifyChannelWebhookOther {
		t.Fatalf("custom webhook capability = %s", got)
	}
	if got := ChannelCapabilityOf("SMS", "SMS_TENCENT"); got != "" {
		t.Fatalf("tencent sms must not be configurable, got capability %s", got)
	}
	if got := ChannelAppsOf("WEBHOOK", []string{CapabilityNotifyChannelWebhookDingTalk}); len(got) != 1 || got[0] != "DINGTALK" {
		t.Fatalf("authorized webhook apps = %v, want [DINGTALK]", got)
	}
	if got := TemplateCapabilityOf("SMS_ALICLOUD", "OTHER"); got != CapabilityNotifyChannelSMSAliCloud {
		t.Fatalf("sms template capability = %s", got)
	}
	if got := TemplateScopesOf([]string{CapabilityNotifyChannelSMSTencent}); len(got) != 0 {
		t.Fatalf("tencent sms should have no template scope, got %v", got)
	}
	if !IsCapability(CapabilityNotifyChannelSMSTencent) {
		t.Fatal("tencent sms capability must stay in the catalog")
	}
	baseModules := set(BaseFeatures())
	for _, code := range base {
		item, _ := CapabilityOf(code)
		if _, ok := baseModules[item.Module]; !ok {
			t.Fatalf("base capability %s module %s is not a base feature", code, item.Module)
		}
	}
	if got := CapabilityPermissions([]string{CapabilityNotifyChannelSMTP, CapabilityNotifyGroup}); len(got) != 6 {
		t.Fatalf("notification permissions = %v, want 3 unique", got)
	}
	if got := CapabilityCodesOfModule(FeatureCluster); len(got) != 24 {
		t.Fatalf("cluster capabilities = %d, want 24", len(got))
	}
}

// TestEffectiveCapabilities 确认套餐能力与额外能力求并集并按目录顺序返回，
// 目录外的旧编码被忽略。
func TestEffectiveCapabilities(t *testing.T) {
	got := EffectiveCapabilities(
		[]string{CapabilityRole, "alert", CapabilityAlertRealtime},
		[]string{CapabilityRole, CapabilityK8sPodRead},
	)
	want := []string{CapabilityAlertRealtime, CapabilityRole, CapabilityK8sPodRead}
	if len(got) != len(want) {
		t.Fatalf("effective capabilities = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("effective capabilities = %v, want %v", got, want)
		}
	}
	modules := ModulesOfCapabilities([]string{CapabilityK8sPodRead, CapabilityClusterAccess, CapabilityMember})
	if len(modules) != 2 || modules[0] != FeatureMember || modules[1] != FeatureCluster {
		t.Fatalf("modules = %v, want [member cluster]", modules)
	}
}

// TestNormalizeCapabilities 确认历史能力码在迁移和兼容路径上展开成当前能力：
// 拆分前的 Webhook / 短信类型编码展开成应用与提供商能力，模块码展开成模块全部能力，
// 未知编码被丢弃，结果按能力目录顺序去重。
func TestNormalizeCapabilities(t *testing.T) {
	got := NormalizeCapabilities([]string{
		"notify_channel_webhook",
		"notify_channel_sms",
		CapabilityNotifyChannelSMTP,
		CapabilityAlertRealtime,
		"unknown",
	})
	want := []string{
		CapabilityAlertRealtime,
		CapabilityNotifyChannelSMTP,
		CapabilityNotifyChannelWebhookOther,
		CapabilityNotifyChannelWebhookDingTalk,
		CapabilityNotifyChannelWebhookWeChat,
		CapabilityNotifyChannelWebhookFeishu,
		CapabilityNotifyChannelSMSAliCloud,
		CapabilityNotifyChannelSMSTencent,
	}
	if !slices.Equal(got, want) {
		t.Fatalf("normalized capabilities = %v, want %v", got, want)
	}
	if got := NormalizeCapabilities([]string{FeatureMember}); len(got) != 1 || got[0] != CapabilityMember {
		t.Fatalf("module code normalization = %v, want [%s]", got, CapabilityMember)
	}
	if !IsLegacyCapability("notify_channel_webhook") || !IsLegacyCapability("notify_channel") {
		t.Fatal("legacy channel codes must be recognized")
	}
	if IsLegacyCapability(CapabilityNotifyChannelSMTP) || IsLegacyCapability("unknown") {
		t.Fatal("current capability or unknown code must not be legacy")
	}
	if got := ExpandLegacyCapability("notify_channel_sms"); !slices.Equal(got, []string{
		CapabilityNotifyChannelSMSAliCloud, CapabilityNotifyChannelSMSTencent,
	}) {
		t.Fatalf("expanded sms capabilities = %v", got)
	}
	if got := ExpandLegacyCapability(CapabilityMember); got != nil {
		t.Fatalf("current capability expanded to %v, want nil", got)
	}
	// M25：旧 Kubernetes 资源能力展开成操作能力，旧 k8s_pod 不含 Pod 终端。
	if got := ExpandLegacyCapability("k8s_pod"); !slices.Equal(got, []string{
		CapabilityK8sPodRead, CapabilityK8sPodLogs, CapabilityK8sPodDelete,
	}) {
		t.Fatalf("expanded k8s_pod capabilities = %v", got)
	}
	if got := ExpandLegacyCapability("k8s_workload"); !slices.Equal(got, []string{
		CapabilityK8sWorkloadRead, CapabilityK8sWorkloadScale, CapabilityK8sWorkloadRestart,
		CapabilityK8sWorkloadImage, CapabilityK8sWorkloadRollback,
	}) {
		t.Fatalf("expanded k8s_workload capabilities = %v", got)
	}
	if got := NormalizeCapabilities([]string{"k8s_cronjob"}); !slices.Equal(got, []string{
		CapabilityK8sCronJobRead, CapabilityK8sCronJobSuspend, CapabilityK8sCronJobTrigger, CapabilityK8sCronJobDelete,
	}) {
		t.Fatalf("normalized k8s_cronjob capabilities = %v", got)
	}
	// 旧写权限展开成操作权限，Pod 终端不随 pod:write 授予。
	if got := ExpandLegacyPermission("k8s:pod:write"); !slices.Equal(got, []string{K8sPodDelete}) {
		t.Fatalf("expanded pod write permission = %v", got)
	}
	if got := ExpandLegacyPermission("k8s:workload:write"); !slices.Equal(got, []string{
		K8sWorkloadScale, K8sWorkloadRestart, K8sWorkloadImage, K8sWorkloadRollback,
	}) {
		t.Fatalf("expanded workload write permission = %v", got)
	}
	if IsLegacyPermission(K8sPodDelete) || !IsLegacyPermission("k8s:pod:write") {
		t.Fatal("legacy permission recognition is wrong")
	}
}

// TestLegacyModuleCapabilities 确认旧模块级编码展开时不会带出 M25 新增的 Pod 终端：
// 存量租户升级后仍要由超级管理员显式开通 k8s_pod_exec。
func TestLegacyModuleCapabilities(t *testing.T) {
	codes := LegacyModuleCapabilities(FeatureCluster)
	if !slices.Contains(codes, CapabilityClusterAccess) || !slices.Contains(codes, CapabilityK8sWorkloadRead) {
		t.Fatalf("legacy cluster capabilities missing known capabilities: %v", codes)
	}
	if slices.Contains(codes, CapabilityK8sPodExec) {
		t.Fatal("legacy module expansion must not grant pod exec")
	}
	if len(codes) != len(CapabilityCodesOfModule(FeatureCluster))-1 {
		t.Fatalf("legacy cluster capabilities = %d, want module capabilities minus pod exec", len(codes))
	}
	if got := LegacyModuleCapabilities(FeatureSSH); !slices.Equal(got, CapabilityCodesOfModule(FeatureSSH)) {
		t.Fatalf("non-cluster module expansion changed: %v", got)
	}
}

func set(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}
