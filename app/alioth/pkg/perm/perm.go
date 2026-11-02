package perm

import "strings"

const (
	MemberRead   = "member:read"
	MemberWrite  = "member:write"
	MemberDelete = "member:delete"

	RoleRead   = "role:read"
	RoleWrite  = "role:write"
	RoleDelete = "role:delete"

	// 大盘是把 Grafana 面板链接组合成整板的展示配置，按读写删拆分权限。
	BoardRead   = "board:read"
	BoardWrite  = "board:write"
	BoardDelete = "board:delete"

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

	// 状态页与 SLA 报表是探测模块下的两个独立菜单。
	StatusPageRead   = "status_page:read"
	StatusPageWrite  = "status_page:write"
	StatusPageDelete = "status_page:delete"
	ProbeSlaRead     = "probe_sla:read"

	AlertRuleRead   = "alert_rule:read"
	AlertRuleWrite  = "alert_rule:write"
	AlertRuleDelete = "alert_rule:delete"

	// 告警按菜单拆分：实时、历史、告警页面、字典各自持有独立权限，
	// 在某个菜单下授权的写/删除不会扩散到同模块的其他菜单。
	AlertRealtimeRead      = "alert_realtime:read"
	AlertRealtimeIntervene = "alert_realtime:intervene"
	AlertHistoryRead       = "alert_history:read"
	AlertPageRead          = "alert_page:read"
	AlertPageWrite         = "alert_page:write"
	AlertPageDelete        = "alert_page:delete"

	// 维护窗口是告警模块下的独立菜单：计划维护期间静默匹配的告警。
	AlertMaintenanceRead   = "alert_maintenance:read"
	AlertMaintenanceWrite  = "alert_maintenance:write"
	AlertMaintenanceDelete = "alert_maintenance:delete"

	// 升级策略是告警模块下的独立菜单：未认领的 FIRING 事件按分钟数逐步升级到通知组。
	AlertEscalationRead   = "alert_escalation:read"
	AlertEscalationWrite  = "alert_escalation:write"
	AlertEscalationDelete = "alert_escalation:delete"

	AlertDictionaryRead   = "alert_dictionary:read"
	AlertDictionaryWrite  = "alert_dictionary:write"
	AlertDictionaryDelete = "alert_dictionary:delete"

	// 通知按菜单拆分：渠道、模板、通知组、订阅、OnCall、消息日志互不共享动作。
	NotificationChannelRead   = "notification_channel:read"
	NotificationChannelWrite  = "notification_channel:write"
	NotificationChannelDelete = "notification_channel:delete"

	NotificationTemplateRead   = "notification_template:read"
	NotificationTemplateWrite  = "notification_template:write"
	NotificationTemplateDelete = "notification_template:delete"

	NotificationGroupRead   = "notification_group:read"
	NotificationGroupWrite  = "notification_group:write"
	NotificationGroupDelete = "notification_group:delete"

	NotificationSubscriptionRead   = "notification_subscription:read"
	NotificationSubscriptionWrite  = "notification_subscription:write"
	NotificationSubscriptionDelete = "notification_subscription:delete"

	NotificationOncallRead   = "notification_oncall:read"
	NotificationOncallWrite  = "notification_oncall:write"
	NotificationOncallDelete = "notification_oncall:delete"

	// 值班轮换是通知模块下的独立菜单：按周期与时区计算当前当班成员。
	NotificationRotationRead   = "notification_rotation:read"
	NotificationRotationWrite  = "notification_rotation:write"
	NotificationRotationDelete = "notification_rotation:delete"

	MessageLogRead   = "message_log:read"
	MessageLogWrite  = "message_log:write"
	MessageLogDelete = "message_log:delete"

	ClusterRead   = "cluster:read"
	ClusterWrite  = "cluster:write"
	ClusterDelete = "cluster:delete"

	// Kubernetes 控制台按「资源 + 操作」细分。cluster:* 只保留集群接入与连接：
	// 列表、详情、测试连通性、修改 kubeconfig 和删除集群；进入控制台仍然要求 cluster:read。
	// 一个操作一个权限，套餐能力也按操作拆：读、封锁节点、扩缩容、滚动重启、设置镜像、
	// 回滚、删除 Pod、Pod 日志、Pod 终端、删除 Job、CronJob 暂停/触发/删除互不共享。
	K8sNodeRead                  = "k8s:node:read"
	K8sNodeCordon                = "k8s:node:cordon"
	K8sWorkloadRead              = "k8s:workload:read"
	K8sWorkloadScale             = "k8s:workload:scale"
	K8sWorkloadRestart           = "k8s:workload:restart"
	K8sWorkloadImage             = "k8s:workload:image"
	K8sWorkloadRollback          = "k8s:workload:rollback"
	K8sPodRead                   = "k8s:pod:read"
	K8sPodLogs                   = "k8s:pod:logs"
	K8sPodDelete                 = "k8s:pod:delete"
	K8sPodExec                   = "k8s:pod:exec"
	K8sIngressRead               = "k8s:ingress:read"
	K8sServiceRead               = "k8s:service:read"
	K8sConfigMapRead             = "k8s:configmap:read"
	K8sSecretRead                = "k8s:secret:read"
	K8sPersistentVolumeRead      = "k8s:pv:read"
	K8sPersistentVolumeClaimRead = "k8s:pvc:read"
	K8sJobRead                   = "k8s:job:read"
	K8sJobDelete                 = "k8s:job:delete"
	K8sCronJobRead               = "k8s:cronjob:read"
	K8sCronJobSuspend            = "k8s:cronjob:suspend"
	K8sCronJobTrigger            = "k8s:cronjob:trigger"
	K8sCronJobDelete             = "k8s:cronjob:delete"

	AgentRead   = "agent:read"
	AgentWrite  = "agent:write"
	AgentDelete = "agent:delete"

	SSHRead   = "ssh:read"
	SSHWrite  = "ssh:write"
	SSHDelete = "ssh:delete"
	SSHExec   = "ssh:exec"
)

// NotificationReadPermissions 返回通知模块所有菜单的读权限。
// 渠道、模板、通知组等元数据会被其他通知菜单引用，读接口允许任一通知菜单只读访问；
// 写和删除仍然按各自菜单独立授权。
func NotificationReadPermissions() []string {
	return []string{
		NotificationChannelRead,
		NotificationTemplateRead,
		NotificationGroupRead,
		NotificationSubscriptionRead,
		NotificationOncallRead,
		MessageLogRead,
	}
}

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
	// 升级与迁移：查看环境状态给平台只读角色；导出 / 导入迁移包和执行数据库迁移
	// 会读写整库数据，只授予超级管理员。
	SystemMigrationRead  = "system:migration:read"
	SystemMigrationWrite = "system:migration:write"
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

type Item struct {
	Name     string
	Resource string
	Action   string
}

// Catalog 返回租户权限目录。权限名全部来自菜单目录：每个菜单声明自己的子权限，
// 因此一个权限只属于一个菜单，不再需要按资源共享资源来归类。
func Catalog() []Item {
	items := make([]Item, 0, len(capabilitySpecs)*3)
	seen := make(map[string]struct{}, len(capabilitySpecs)*3)
	for _, menu := range menuSpecs {
		for _, name := range menu.Permissions {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			items = append(items, itemOf(name))
		}
	}
	items = append(items, menuCatalog()...)
	return items
}

// itemOf 把权限名拆成资源与动作：k8s:node:read -> resource k8s:node, action read。
func itemOf(name string) Item {
	index := strings.LastIndex(name, ":")
	if index < 0 {
		return Item{Name: name, Resource: name, Action: name}
	}
	return Item{Name: name, Resource: name[:index], Action: name[index+1:]}
}

// menuCatalog 返回全部菜单权限，顺序与菜单目录一致。
func menuCatalog() []Item {
	items := make([]Item, 0, len(menuSpecs))
	for _, item := range menuSpecs {
		items = append(items, Item{Name: MenuPermission(item.Code), Resource: "menu", Action: item.Code})
	}
	return items
}

// k8sCatalog 返回 Kubernetes「资源 + 操作」级权限，追加在租户权限目录末尾。
// Resource 用 k8s:<资源> 形式，供前端按资源分组；Name 仍是 Resource + ":" + Action。
func k8sCatalog() []Item {
	return []Item{
		{Name: K8sNodeRead, Resource: "k8s:node", Action: "read"},
		{Name: K8sNodeCordon, Resource: "k8s:node", Action: "cordon"},
		{Name: K8sWorkloadRead, Resource: "k8s:workload", Action: "read"},
		{Name: K8sWorkloadScale, Resource: "k8s:workload", Action: "scale"},
		{Name: K8sWorkloadRestart, Resource: "k8s:workload", Action: "restart"},
		{Name: K8sWorkloadImage, Resource: "k8s:workload", Action: "image"},
		{Name: K8sWorkloadRollback, Resource: "k8s:workload", Action: "rollback"},
		{Name: K8sPodRead, Resource: "k8s:pod", Action: "read"},
		{Name: K8sPodLogs, Resource: "k8s:pod", Action: "logs"},
		{Name: K8sPodDelete, Resource: "k8s:pod", Action: "delete"},
		{Name: K8sPodExec, Resource: "k8s:pod", Action: "exec"},
		{Name: K8sIngressRead, Resource: "k8s:ingress", Action: "read"},
		{Name: K8sServiceRead, Resource: "k8s:service", Action: "read"},
		{Name: K8sConfigMapRead, Resource: "k8s:configmap", Action: "read"},
		{Name: K8sSecretRead, Resource: "k8s:secret", Action: "read"},
		{Name: K8sPersistentVolumeRead, Resource: "k8s:pv", Action: "read"},
		{Name: K8sPersistentVolumeClaimRead, Resource: "k8s:pvc", Action: "read"},
		{Name: K8sJobRead, Resource: "k8s:job", Action: "read"},
		{Name: K8sJobDelete, Resource: "k8s:job", Action: "delete"},
		{Name: K8sCronJobRead, Resource: "k8s:cronjob", Action: "read"},
		{Name: K8sCronJobSuspend, Resource: "k8s:cronjob", Action: "suspend"},
		{Name: K8sCronJobTrigger, Resource: "k8s:cronjob", Action: "trigger"},
		{Name: K8sCronJobDelete, Resource: "k8s:cronjob", Action: "delete"},
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
		{Name: SystemMigrationRead, Resource: "system", Action: "migration:read"},
		{Name: SystemMigrationWrite, Resource: "system", Action: "migration:write"},
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

// SystemPermissions 返回平台角色拥有的系统权限。
// 系统权限不叠加租户权限；平台管理员跨租户只读，写操作仍要求成为该租户成员。
func SystemPermissions(role string) []string {
	switch role {
	case SystemRoleSuperAdmin:
		return SystemNames()
	case SystemRoleSystemAdmin:
		return []string{
			SystemUserRead, SystemUserWrite, SystemRoleRead, SystemInviteRead,
			SystemNamespaceRead, SystemPackageRead, SystemMigrationRead,
		}
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
// Pod 日志读取（k8s:pod:logs）不改变集群状态，也属于只读；平台管理员只读视角可读。
func K8sReadPermissions() []string {
	names := make([]string, 0, len(k8sCatalog()))
	for _, item := range k8sCatalog() {
		if item.Action == "read" || item.Action == "logs" {
			names = append(names, item.Name)
		}
	}
	return names
}

// K8sWritePermissions 返回 Kubernetes 资源级的全部操作权限（只读之外）。
// 迁移旧的自定义角色时，持有 cluster:write 的角色补齐这一组。
// pod:exec 是 M25 新增的独立权限，不在这里：它不随任何旧能力或旧 cluster:write 自动授予，
// 必须由超级管理员在套餐里显式开通 k8s_pod_exec。
func K8sWritePermissions() []string {
	names := make([]string, 0, len(k8sCatalog()))
	for _, item := range k8sCatalog() {
		if item.Name == K8sPodExec {
			continue
		}
		if item.Action != "read" && item.Action != "logs" {
			names = append(names, item.Name)
		}
	}
	return names
}

// legacyK8sPermissionExpansion 是 M25 之前的「资源写权限」到「操作权限」的展开表。
// 旧角色持有 k8s:*:write 时按这张表补齐操作权限，迁移完成后旧权限行与角色关联被清理。
// Pod Exec 不在表里：它是 M25 新增权限，旧 write 不隐含容器终端。
var legacyK8sPermissionExpansion = map[string][]string{
	"k8s:node:write":     {K8sNodeCordon},
	"k8s:workload:write": {K8sWorkloadScale, K8sWorkloadRestart, K8sWorkloadImage, K8sWorkloadRollback},
	"k8s:pod:write":      {K8sPodDelete},
	"k8s:job:write":      {K8sJobDelete},
	"k8s:cronjob:write":  {K8sCronJobSuspend, K8sCronJobTrigger, K8sCronJobDelete},
}

// ExpandLegacyPermission 返回旧权限名对应的当前权限；不是旧权限时返回 nil。
func ExpandLegacyPermission(name string) []string {
	codes, ok := legacyK8sPermissionExpansion[name]
	if !ok {
		return nil
	}
	out := make([]string, len(codes))
	copy(out, codes)
	return out
}

// IsLegacyPermission 判断权限名是不是 M25 之前的历史写权限。
func IsLegacyPermission(name string) bool {
	_, ok := legacyK8sPermissionExpansion[name]
	return ok
}
