package perm

import "strings"

const (
	MemberRead   = "member:read"
	MemberWrite  = "member:write"
	MemberDelete = "member:delete"

	RoleRead   = "role:read"
	RoleWrite  = "role:write"
	RoleDelete = "role:delete"

	NamespaceRead     = "namespace:read"
	NamespaceWrite    = "namespace:write"
	NamespaceDelete   = "namespace:delete"
	NamespaceTransfer = "namespace:transfer"

	DatasourceRead   = "datasource:read"
	DatasourceWrite  = "datasource:write"
	DatasourceDelete = "datasource:delete"

	// Probe 是探测目标组的权限点。探测规则本身属于告警规则，用 alert_rule:*。
	ProbeRead   = "probe:read"
	ProbeWrite  = "probe:write"
	ProbeDelete = "probe:delete"

	AlertRuleRead   = "alert_rule:read"
	AlertRuleWrite  = "alert_rule:write"
	AlertRuleDelete = "alert_rule:delete"

	AlertRead      = "alert:read"
	AlertWrite     = "alert:write"
	AlertDelete    = "alert:delete"
	AlertIntervene = "alert:intervene"

	NotificationRead   = "notification:read"
	NotificationWrite  = "notification:write"
	NotificationDelete = "notification:delete"

	ClusterRead   = "cluster:read"
	ClusterWrite  = "cluster:write"
	ClusterDelete = "cluster:delete"

	// Kubernetes 控制台按资源细分。cluster:* 只保留集群接入与连接：列表、详情、
	// 测试连通性、修改 kubeconfig 和删除集群；进入控制台仍然要求 cluster:read。
	// 节点、工作负载、Pod 等资源的读写用下面的 k8s:<资源>:<动作>。Pod 日志单独授权。
	K8sNodeRead                  = "k8s:node:read"
	K8sNodeWrite                 = "k8s:node:write"
	K8sWorkloadRead              = "k8s:workload:read"
	K8sWorkloadWrite             = "k8s:workload:write"
	K8sPodRead                   = "k8s:pod:read"
	K8sPodWrite                  = "k8s:pod:write"
	K8sPodLogs                   = "k8s:pod:logs"
	K8sIngressRead               = "k8s:ingress:read"
	K8sServiceRead               = "k8s:service:read"
	K8sConfigMapRead             = "k8s:configmap:read"
	K8sSecretRead                = "k8s:secret:read"
	K8sPersistentVolumeRead      = "k8s:pv:read"
	K8sPersistentVolumeClaimRead = "k8s:pvc:read"
	K8sJobRead                   = "k8s:job:read"
	K8sJobWrite                  = "k8s:job:write"
	K8sCronJobRead               = "k8s:cronjob:read"
	K8sCronJobWrite              = "k8s:cronjob:write"

	AgentRead   = "agent:read"
	AgentWrite  = "agent:write"
	AgentDelete = "agent:delete"

	SSHRead   = "ssh:read"
	SSHWrite  = "ssh:write"
	SSHDelete = "ssh:delete"
	SSHExec   = "ssh:exec"
)

// 系统权限属于平台，不写入租户的 permission 表，也不参与租户角色授权。
const (
	SystemUserRead      = "system:user:read"
	SystemUserWrite     = "system:user:write"
	SystemRoleWrite     = "system:role:write"
	SystemRoleRead      = "system:role:read"
	SystemInviteRead    = "system:invite:read"
	SystemNamespaceRead = "system:namespace:read"
	// system:namespace:write 管理租户生命周期：停用、启用和归档，只给超级管理员。
	SystemNamespaceWrite = "system:namespace:write"
	// 功能套餐只由平台维护：读给系统管理员，写只给超级管理员。
	SystemPackageRead  = "system:package:read"
	SystemPackageWrite = "system:package:write"
)

// 套餐版本决定套餐出现在哪里：公开套餐可被用户在创建租户时选择，
// 内测套餐只由平台管理员在租户功能授权里分配。
const (
	PackageVisibilityPublic = "PUBLIC"
	PackageVisibilityBeta   = "BETA"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// 平台角色只有固定四档。第一个注册用户是超级管理员，其余由超级管理员授权。
// SYSTEM_READONLY 是可查看平台全部信息、但没有任何写权限的只读管理员。
const (
	SystemRoleSuperAdmin  = "SUPER_ADMIN"
	SystemRoleSystemAdmin = "SYSTEM_ADMIN"
	SystemRoleReadOnly    = "SYSTEM_READONLY"
	SystemRoleUser        = "USER"
)

// resources 只包含租户资源。user 是平台资源，不在租户权限目录里。
var resources = []string{
	"member", "role", "namespace", "datasource", "probe", "alert_rule",
	"alert", "notification", "cluster", "agent", "ssh",
}

type Item struct {
	Name     string
	Resource string
	Action   string
}

// Catalog 返回租户权限目录。
func Catalog() []Item {
	items := make([]Item, 0, len(resources)*2+4+len(k8sCatalog())+len(capabilitySpecs))
	for _, resource := range resources {
		items = append(items,
			Item{Name: resource + ":read", Resource: resource, Action: "read"},
			Item{Name: resource + ":write", Resource: resource, Action: "write"},
		)
		// 删除普通业务数据是可以单独授予的权限；删除租户是另一回事，
		// namespace:delete 只属于 owner，仍在下面单独列出。
		if resource != "namespace" {
			items = append(items, Item{Name: resource + ":delete", Resource: resource, Action: "delete"})
		}
	}
	items = append(items,
		Item{Name: AlertIntervene, Resource: "alert", Action: "intervene"},
		Item{Name: SSHExec, Resource: "ssh", Action: "exec"},
		Item{Name: NamespaceDelete, Resource: "namespace", Action: "delete"},
		Item{Name: NamespaceTransfer, Resource: "namespace", Action: "transfer"},
	)
	items = append(items, k8sCatalog()...)
	// 菜单权限追加在租户权限目录末尾。资源名固定为 menu，动作是能力编码，
	// 因此 Name 仍是 Resource+":"+Action 的形式，前端可以直接按资源分组。
	items = append(items, menuCatalog()...)
	return items
}

// menuCatalog 返回全部能力的菜单权限，顺序与能力目录一致。
func menuCatalog() []Item {
	items := make([]Item, 0, len(capabilitySpecs))
	for _, item := range capabilitySpecs {
		items = append(items, Item{Name: MenuPermission(item.Code), Resource: "menu", Action: item.Code})
	}
	return items
}

// k8sCatalog 返回 Kubernetes 资源级权限，追加在租户权限目录末尾。
// Resource 用 k8s:<资源> 形式，供前端按资源分组；Name 仍是 Resource + ":" + Action。
func k8sCatalog() []Item {
	return []Item{
		{Name: K8sNodeRead, Resource: "k8s:node", Action: "read"},
		{Name: K8sNodeWrite, Resource: "k8s:node", Action: "write"},
		{Name: K8sWorkloadRead, Resource: "k8s:workload", Action: "read"},
		{Name: K8sWorkloadWrite, Resource: "k8s:workload", Action: "write"},
		{Name: K8sPodRead, Resource: "k8s:pod", Action: "read"},
		{Name: K8sPodWrite, Resource: "k8s:pod", Action: "write"},
		{Name: K8sPodLogs, Resource: "k8s:pod", Action: "logs"},
		{Name: K8sIngressRead, Resource: "k8s:ingress", Action: "read"},
		{Name: K8sServiceRead, Resource: "k8s:service", Action: "read"},
		{Name: K8sConfigMapRead, Resource: "k8s:configmap", Action: "read"},
		{Name: K8sSecretRead, Resource: "k8s:secret", Action: "read"},
		{Name: K8sPersistentVolumeRead, Resource: "k8s:pv", Action: "read"},
		{Name: K8sPersistentVolumeClaimRead, Resource: "k8s:pvc", Action: "read"},
		{Name: K8sJobRead, Resource: "k8s:job", Action: "read"},
		{Name: K8sJobWrite, Resource: "k8s:job", Action: "write"},
		{Name: K8sCronJobRead, Resource: "k8s:cronjob", Action: "read"},
		{Name: K8sCronJobWrite, Resource: "k8s:cronjob", Action: "write"},
	}
}

// SystemCatalog 返回平台权限目录，和租户权限目录互不混用。
func SystemCatalog() []Item {
	return []Item{
		{Name: SystemUserRead, Resource: "system", Action: "user:read"},
		{Name: SystemUserWrite, Resource: "system", Action: "user:write"},
		{Name: SystemRoleRead, Resource: "system", Action: "role:read"},
		{Name: SystemRoleWrite, Resource: "system", Action: "role:write"},
		{Name: SystemInviteRead, Resource: "system", Action: "invite:read"},
		{Name: SystemNamespaceRead, Resource: "system", Action: "namespace:read"},
		{Name: SystemNamespaceWrite, Resource: "system", Action: "namespace:write"},
		{Name: SystemPackageRead, Resource: "system", Action: "package:read"},
		{Name: SystemPackageWrite, Resource: "system", Action: "package:write"},
	}
}

func Names() []string {
	catalog := Catalog()
	names := make([]string, len(catalog))
	for i, item := range catalog {
		names[i] = item.Name
	}
	return names
}

func SystemNames() []string {
	catalog := SystemCatalog()
	names := make([]string, len(catalog))
	for i, item := range catalog {
		names[i] = item.Name
	}
	return names
}

func RolePermissions(role string) []string {
	switch role {
	case RoleOwner:
		return Names()
	case RoleAdmin:
		return without(Names(), NamespaceDelete, NamespaceTransfer)
	case RoleMember:
		return memberPermissions()
	case RoleViewer:
		return readPermissions()
	default:
		return nil
	}
}

// DefaultRoleRemark 返回预置角色的默认备注，是新租户建角色和 seed 补空备注的唯一来源。
// 备注是展示给管理员看的说明，不参与鉴权判断。
func DefaultRoleRemark(role string) string {
	switch role {
	case RoleOwner:
		return "租户负责人，拥有全部权限，包括删除租户与移交负责人。"
	case RoleAdmin:
		return "可管理成员、角色与配置，不能删除租户或移交负责人。"
	case RoleMember:
		return "日常运维：查看全部资源，处理告警与 SSH 执行。"
	case RoleViewer:
		return "只读：可以查看但不能修改任何配置。"
	default:
		return ""
	}
}

// SystemPermissions 返回平台角色拥有的系统权限。
// 系统权限不叠加租户权限；平台管理员跨租户只读，写操作仍要求成为该租户成员。
func SystemPermissions(role string) []string {
	switch role {
	case SystemRoleSuperAdmin:
		return SystemNames()
	case SystemRoleSystemAdmin:
		return []string{SystemUserRead, SystemUserWrite, SystemRoleRead, SystemInviteRead, SystemNamespaceRead, SystemPackageRead}
	case SystemRoleReadOnly:
		return systemReadPermissions()
	default:
		return nil
	}
}

// systemReadPermissions 返回系统权限目录里的全部只读权限。
// 系统只读角色靠它查看平台信息，写权限一个都不持有。
func systemReadPermissions() []string {
	names := make([]string, 0, len(SystemCatalog()))
	for _, item := range SystemCatalog() {
		if item.Action == "read" || strings.HasSuffix(item.Name, ":read") {
			names = append(names, item.Name)
		}
	}
	return names
}

// IsReadOnly 判断权限名是否属于租户目录中的只读权限。
// 只读动作是 read 和 logs：读取 Pod 日志不改变集群状态，但仍然是独立权限点。
// 平台管理员只靠它跨租户读取；写、删除、介入和执行动作一律不放行。
func IsReadOnly(code string) bool {
	for _, item := range Catalog() {
		if item.Name != code {
			continue
		}
		return item.Resource == "menu" || item.Action == "read" || item.Action == "logs"
	}
	return false
}

// ReadOnlyNames 返回租户目录中的全部只读权限名，顺序与 Catalog 一致。
// 平台管理员（含系统只读角色）不是租户成员时，/self/permissions 返回这份目录。
// 菜单权限只是入口开关，不改变任何状态，也算只读。
func ReadOnlyNames() []string {
	names := make([]string, 0)
	for _, item := range Catalog() {
		if item.Resource == "menu" || item.Action == "read" || item.Action == "logs" {
			names = append(names, item.Name)
		}
	}
	return names
}

// TenantRoles 返回每个租户预置的四个角色。
func TenantRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleMember, RoleViewer}
}

// SystemRoles 返回平台固定角色。
func SystemRoles() []string {
	return []string{SystemRoleSuperAdmin, SystemRoleSystemAdmin, SystemRoleReadOnly, SystemRoleUser}
}

// IsTenantPermission 只认租户权限目录。用于阻止把 user:* 或 system:* 授给租户角色。
func IsTenantPermission(name string) bool {
	for _, item := range Catalog() {
		if item.Name == name {
			return true
		}
	}
	return false
}

func IsSystemRole(role string) bool {
	switch role {
	case SystemRoleSuperAdmin, SystemRoleSystemAdmin, SystemRoleReadOnly, SystemRoleUser:
		return true
	default:
		return false
	}
}

// IsSystemReadOnly 判断平台角色是否是只读管理员。
// 只读管理员能看平台全部信息，但任何写接口都必须返回 403。
func IsSystemReadOnly(role string) bool {
	return role == SystemRoleReadOnly
}

// CanReadAnyTenant 判断平台角色是否可以跨租户只读。
// 只读管理员只是不写，读取范围与超级管理员、系统管理员一致。
func CanReadAnyTenant(role string) bool {
	switch role {
	case SystemRoleSuperAdmin, SystemRoleSystemAdmin, SystemRoleReadOnly:
		return true
	default:
		return false
	}
}

func IsSystemAdmin(role string) bool {
	return role == SystemRoleSuperAdmin || role == SystemRoleSystemAdmin
}

// K8sReadPermissions 返回 Kubernetes 资源级的全部只读权限，顺序与 Catalog 一致。
// Pod 日志读取（k8s:pod:logs）不改变集群状态，也属于只读；member 和 viewer 默认持有。
func K8sReadPermissions() []string {
	names := make([]string, 0, len(k8sCatalog()))
	for _, item := range k8sCatalog() {
		if item.Action == "read" || item.Action == "logs" {
			names = append(names, item.Name)
		}
	}
	return names
}

// K8sWritePermissions 返回 Kubernetes 资源级的全部写权限。
// 迁移旧的自定义角色时，持有 cluster:write 的角色补齐这一组，但不会因此拿到 pod:logs。
func K8sWritePermissions() []string {
	names := make([]string, 0, len(k8sCatalog()))
	for _, item := range k8sCatalog() {
		if item.Action == "write" {
			names = append(names, item.Name)
		}
	}
	return names
}

// readPermissions 返回 viewer 的只读权限：全部租户资源的读权限，加上 Kubernetes
// 资源级只读权限（含读取 Pod 日志）。member 在此基础上再叠加告警干预与 SSH 执行。
func readPermissions() []string {
	names := make([]string, 0, len(resources)+len(k8sCatalog())+len(capabilitySpecs))
	for _, resource := range resources {
		names = append(names, resource+":read")
	}
	names = append(names, K8sReadPermissions()...)
	// 菜单权限只控制入口可见性，属于只读：viewer 与 member 默认持有全部菜单，
	// 实际可见哪些入口仍由套餐开通的能力过滤。
	return append(names, MenuPermissions()...)
}

func memberPermissions() []string {
	names := readPermissions()
	return append(names, AlertIntervene, SSHExec)
}

func without(names []string, drop ...string) []string {
	blocked := map[string]struct{}{}
	for _, name := range drop {
		blocked[name] = struct{}{}
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := blocked[name]; ok {
			continue
		}
		out = append(out, name)
	}
	return out
}
