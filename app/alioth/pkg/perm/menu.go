package perm

import "slices"

// 菜单是角色授权的单位，对应侧边栏的一个可授权入口；能力只负责套餐开通。
// 一个菜单可以由多个能力开通（数据源六种驱动、通知渠道三种类型），
// 菜单下的权限是角色实际可勾选的子权限，勾选菜单默认包含它的读权限。
type Menu struct {
	Code         string
	Group        string
	Name         string
	Capabilities []string
	Permissions  []string
	DefaultReads []string
	// Always 表示不受套餐能力限制，始终可用（namespace:* 的租户设置菜单）。
	Always bool
}

// menuSpecs 是菜单目录的唯一来源。顺序即套餐勾选页与角色授权页的展示顺序。
var menuSpecs = []Menu{
	{Code: "alert_realtime", Group: FeatureAlert, Name: "实时告警", Capabilities: []string{CapabilityAlertRealtime}, Permissions: []string{AlertRead, AlertIntervene}, DefaultReads: []string{AlertRead}},
	{Code: "alert_history", Group: FeatureAlert, Name: "历史告警", Capabilities: []string{CapabilityAlertHistory}, Permissions: []string{AlertRead}, DefaultReads: []string{AlertRead}},
	{Code: "alert_page", Group: FeatureAlert, Name: "告警页面", Capabilities: []string{CapabilityAlertPage}, Permissions: []string{AlertRead, AlertWrite, AlertDelete}, DefaultReads: []string{AlertRead}},
	{Code: "alert_rule", Group: FeatureAlertRule, Name: "告警规则", Capabilities: []string{CapabilityAlertRule}, Permissions: []string{AlertRuleRead, AlertRuleWrite, AlertRuleDelete}, DefaultReads: []string{AlertRuleRead}},
	{Code: "alert_dictionary", Group: FeatureAlertRule, Name: "告警字典", Capabilities: []string{CapabilityAlertDictionary}, Permissions: []string{AlertRuleRead, AlertRuleWrite, AlertRuleDelete}, DefaultReads: []string{AlertRuleRead}},
	{Code: "probe", Group: FeatureProbe, Name: "探测目标", Capabilities: []string{CapabilityProbeHTTP}, Permissions: []string{ProbeRead, ProbeWrite, ProbeDelete}, DefaultReads: []string{ProbeRead}},
	{Code: "datasource", Group: FeatureDatasource, Name: "数据源", Capabilities: DatasourceCapabilities(), Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete}, DefaultReads: []string{DatasourceRead}},
	{Code: "notification_channel", Group: FeatureNotification, Name: "通知渠道", Capabilities: []string{CapabilityNotifyChannelSMTP, CapabilityNotifyChannelWebhook, CapabilityNotifyChannelSMS}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "notification_template", Group: FeatureNotification, Name: "消息模板", Capabilities: []string{CapabilityNotifyTemplate}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "notification_group", Group: FeatureNotification, Name: "通知组", Capabilities: []string{CapabilityNotifyGroup}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "notification_subscription", Group: FeatureNotification, Name: "告警订阅", Capabilities: []string{CapabilityNotifySubscription}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "notification_oncall", Group: FeatureNotification, Name: "OnCall 值班", Capabilities: []string{CapabilityNotifyOnCall}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "message_log", Group: FeatureNotification, Name: "消息日志", Capabilities: []string{CapabilityMessageLog}, Permissions: []string{NotificationRead, NotificationWrite, NotificationDelete}, DefaultReads: []string{NotificationRead}},
	{Code: "member", Group: FeatureMember, Name: "成员", Capabilities: []string{CapabilityMember}, Permissions: []string{MemberRead, MemberWrite, MemberDelete}, DefaultReads: []string{MemberRead}},
	{Code: "role", Group: FeatureRole, Name: "角色", Capabilities: []string{CapabilityRole}, Permissions: []string{RoleRead, RoleWrite, RoleDelete}, DefaultReads: []string{RoleRead}},
	{Code: "namespace", Group: "tenant", Name: "租户设置", Always: true, Permissions: []string{NamespaceRead, NamespaceWrite, NamespaceDelete, NamespaceTransfer}, DefaultReads: []string{NamespaceRead}},
	{Code: "cluster_access", Group: FeatureCluster, Name: "集群接入", Capabilities: []string{CapabilityClusterAccess}, Permissions: []string{ClusterRead, ClusterWrite, ClusterDelete}, DefaultReads: []string{ClusterRead}},
	{Code: "k8s_node", Group: FeatureCluster, Name: "节点", Capabilities: []string{CapabilityK8sNode}, Permissions: []string{K8sNodeRead, K8sNodeWrite}, DefaultReads: []string{K8sNodeRead}},
	{Code: "k8s_workload", Group: FeatureCluster, Name: "工作负载", Capabilities: []string{CapabilityK8sWorkload}, Permissions: []string{K8sWorkloadRead, K8sWorkloadWrite}, DefaultReads: []string{K8sWorkloadRead}},
	{Code: "k8s_pod", Group: FeatureCluster, Name: "Pod", Capabilities: []string{CapabilityK8sPod}, Permissions: []string{K8sPodRead, K8sPodWrite, K8sPodLogs}, DefaultReads: []string{K8sPodRead, K8sPodLogs}},
	{Code: "k8s_ingress", Group: FeatureCluster, Name: "Ingress", Capabilities: []string{CapabilityK8sIngress}, Permissions: []string{K8sIngressRead}, DefaultReads: []string{K8sIngressRead}},
	{Code: "k8s_service", Group: FeatureCluster, Name: "Service", Capabilities: []string{CapabilityK8sService}, Permissions: []string{K8sServiceRead}, DefaultReads: []string{K8sServiceRead}},
	{Code: "k8s_configmap", Group: FeatureCluster, Name: "ConfigMap", Capabilities: []string{CapabilityK8sConfigMap}, Permissions: []string{K8sConfigMapRead}, DefaultReads: []string{K8sConfigMapRead}},
	{Code: "k8s_secret", Group: FeatureCluster, Name: "Secret", Capabilities: []string{CapabilityK8sSecret}, Permissions: []string{K8sSecretRead}, DefaultReads: []string{K8sSecretRead}},
	{Code: "k8s_pv", Group: FeatureCluster, Name: "PV", Capabilities: []string{CapabilityK8sPV}, Permissions: []string{K8sPersistentVolumeRead}, DefaultReads: []string{K8sPersistentVolumeRead}},
	{Code: "k8s_pvc", Group: FeatureCluster, Name: "PVC", Capabilities: []string{CapabilityK8sPVC}, Permissions: []string{K8sPersistentVolumeClaimRead}, DefaultReads: []string{K8sPersistentVolumeClaimRead}},
	{Code: "k8s_job", Group: FeatureCluster, Name: "Job", Capabilities: []string{CapabilityK8sJob}, Permissions: []string{K8sJobRead, K8sJobWrite}, DefaultReads: []string{K8sJobRead}},
	{Code: "k8s_cronjob", Group: FeatureCluster, Name: "CronJob", Capabilities: []string{CapabilityK8sCronJob}, Permissions: []string{K8sCronJobRead, K8sCronJobWrite}, DefaultReads: []string{K8sCronJobRead}},
	{Code: "agent", Group: FeatureAgent, Name: "机器", Capabilities: []string{CapabilityAgentMachine}, Permissions: []string{AgentRead, AgentWrite, AgentDelete}, DefaultReads: []string{AgentRead}},
	{Code: "ssh_command", Group: FeatureSSH, Name: "SSH 命令", Capabilities: []string{CapabilitySSHCommand}, Permissions: []string{SSHRead, SSHWrite, SSHDelete}, DefaultReads: []string{SSHRead}},
	{Code: "web_shell", Group: FeatureSSH, Name: "Web Shell", Capabilities: []string{CapabilityWebShell}, Permissions: []string{SSHExec}},
}

var (
	menuIndex = func() map[string]Menu {
		index := make(map[string]Menu, len(menuSpecs))
		for _, item := range menuSpecs {
			index[item.Code] = item
		}
		return index
	}()
	menuPermissionIndex = func() map[string]string {
		index := make(map[string]string, len(menuSpecs))
		for _, item := range menuSpecs {
			index[MenuPermission(item.Code)] = item.Code
		}
		return index
	}()
)

// Menus 按固定顺序返回全部菜单。
func Menus() []Menu {
	items := make([]Menu, 0, len(menuSpecs))
	for _, item := range menuSpecs {
		items = append(items, Menu{
			Code: item.Code, Group: item.Group, Name: item.Name, Always: item.Always,
			Capabilities: cloneStrings(item.Capabilities),
			Permissions:  cloneStrings(item.Permissions),
			DefaultReads: cloneStrings(item.DefaultReads),
		})
	}
	return items
}

// MenuOf 返回菜单详情；code 不存在时第二个返回值为 false。
func MenuOf(code string) (Menu, bool) {
	item, ok := menuIndex[code]
	if !ok {
		return Menu{}, false
	}
	item.Capabilities = cloneStrings(item.Capabilities)
	item.Permissions = cloneStrings(item.Permissions)
	item.DefaultReads = cloneStrings(item.DefaultReads)
	return item, true
}

// IsMenu 判断编码是否在菜单目录里。
func IsMenu(code string) bool {
	_, ok := menuIndex[code]
	return ok
}

// MenuPermission 返回菜单对应的菜单权限名。
func MenuPermission(code string) string {
	return "menu:" + code
}

// MenuOfMenuPermission 返回菜单权限对应的菜单编码；不是合法菜单权限时返回空字符串。
func MenuOfMenuPermission(permission string) (string, bool) {
	code, ok := menuPermissionIndex[permission]
	return code, ok
}

// MenuPermissionCapability 保留旧名字，返回菜单权限对应的菜单编码。
func MenuPermissionCapability(permission string) string {
	code, _ := MenuOfMenuPermission(permission)
	return code
}

// IsMenuPermission 判断权限名是不是菜单目录里的菜单权限。
func IsMenuPermission(permission string) bool {
	_, ok := MenuOfMenuPermission(permission)
	return ok
}

// MenuPermissions 按菜单目录顺序返回全部菜单权限名。
func MenuPermissions() []string {
	names := make([]string, 0, len(menuSpecs))
	for _, item := range menuSpecs {
		names = append(names, MenuPermission(item.Code))
	}
	return names
}

// DefaultReadsOfMenu 返回菜单勾选后默认授予的读权限。
func DefaultReadsOfMenu(code string) []string {
	item, ok := MenuOf(code)
	if !ok {
		return nil
	}
	return item.DefaultReads
}

// MenusOfCapability 返回可由该能力开通的菜单编码，按目录顺序。
func MenusOfCapability(capability string) []string {
	codes := make([]string, 0, 2)
	for _, item := range menuSpecs {
		if slices.Contains(item.Capabilities, capability) {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// MenusOfPermission 返回声明该权限的菜单编码，按目录顺序。只读权限可以被多个菜单共享。
func MenusOfPermission(permission string) []string {
	codes := make([]string, 0, 2)
	for _, item := range menuSpecs {
		if slices.Contains(item.Permissions, permission) {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// MenuLicensed 判断菜单是否被这组能力开通；Always 菜单始终可用。
func MenuLicensed(code string, capabilities []string) bool {
	item, ok := MenuOf(code)
	if !ok {
		return false
	}
	if item.Always {
		return true
	}
	for _, capability := range item.Capabilities {
		if slices.Contains(capabilities, capability) {
			return true
		}
	}
	return false
}

// EffectiveMenus 返回被这组能力开通的菜单，按目录顺序。
func EffectiveMenus(capabilities []string) []string {
	codes := make([]string, 0, len(menuSpecs))
	for _, item := range menuSpecs {
		if MenuLicensed(item.Code, capabilities) {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// PermissionLicensed 判断权限是否属于租户已开通的能力范围。
// namespace:* 不属于任何能力，始终可授权；菜单权限单独按 MenuLicensed 判断。
func PermissionLicensed(permission string, capabilities []string) bool {
	if IsMenuPermission(permission) {
		code, _ := MenuOfMenuPermission(permission)
		return MenuLicensed(code, capabilities)
	}
	caps := PermissionCapabilities(permission)
	if len(caps) == 0 {
		return slices.Contains([]string{NamespaceRead, NamespaceWrite, NamespaceDelete, NamespaceTransfer}, permission)
	}
	for _, capability := range caps {
		if slices.Contains(capabilities, capability) {
			return true
		}
	}
	return false
}

func cloneStrings(values []string) []string {
	out := make([]string, len(values))
	copy(out, values)
	return out
}
