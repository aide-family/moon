package perm

import (
	"slices"
)

// 能力是模块下可单独开通的最小授权单位。两级功能模型里：
//   - 模块（Feature）决定菜单分组，是租户差异与功能检查的语义单位；
//   - 能力（Capability）是套餐里勾选的单位，一个模块下可以有多个能力；
//   - 权限（Permission）挂在能力上，租户角色按能力持有的权限决定能调哪些接口。
//
// 套餐与租户差异都只存能力编码；模块由能力反推，不单独落表。
const (
	CapabilityAlertRealtime   = "alert_realtime"
	CapabilityAlertHistory    = "alert_history"
	CapabilityAlertPage       = "alert_page"
	CapabilityAlertRule       = "alert_rule"
	CapabilityAlertDictionary = "alert_dictionary"
	// 探测目标按驱动细分。当前只有 HTTP，TCP / ICMP 在实现探测执行器后新增能力，
	// 旧能力码不再复用，避免套餐把未实现的类型卖给租户。
	CapabilityProbeHTTP = "probe_http"
	// 数据源按驱动细分：一个套餐可以只买 Prometheus，不买 Loki 或 ES。
	// 六个编码与 datasource.driver 枚举一一对应，权限相同，由能力决定可用类型。
	CapabilityDatasourcePrometheus      = "datasource_prometheus"
	CapabilityDatasourceVictoriaMetrics = "datasource_victoria_metrics"
	CapabilityDatasourceElasticsearch   = "datasource_elasticsearch"
	CapabilityDatasourceLoki            = "datasource_loki"
	CapabilityDatasourceVictoriaLogs    = "datasource_vlogs"
	CapabilityDatasourceJaeger          = "datasource_jaeger"
	// 通知模块按菜单拆成七类能力，权限相同，由菜单权限区分入口；
	// 通知渠道再按「渠道类型 + 应用」细分：Webhook 按厂商拆成四个应用能力，
	// 短信按提供商拆成阿里云和腾讯云两个能力。未授权的渠道能力不出现在渠道、
	// 模板和通知组里。腾讯云短信只进目录可售，暂不支持配置和发送。
	CapabilityNotifyChannelSMTP            = "notify_channel_smtp"
	CapabilityNotifyChannelWebhookOther    = "notify_channel_webhook_other"
	CapabilityNotifyChannelWebhookDingTalk = "notify_channel_webhook_dingtalk"
	CapabilityNotifyChannelWebhookWeChat   = "notify_channel_webhook_wechat"
	CapabilityNotifyChannelWebhookFeishu   = "notify_channel_webhook_feishu"
	CapabilityNotifyChannelSMSAliCloud     = "notify_channel_sms_alicloud"
	CapabilityNotifyChannelSMSTencent      = "notify_channel_sms_tencent"
	CapabilityNotifyTemplate               = "notify_template"
	CapabilityNotifyGroup                  = "notify_group"
	CapabilityNotifySubscription           = "notify_subscription"
	CapabilityNotifyOnCall                 = "notify_oncall"
	CapabilityMessageLog                   = "message_log"
	CapabilityMember                       = "member"
	CapabilityRole                         = "role"
	CapabilityClusterAccess                = "cluster_access"
	CapabilityK8sNode                      = "k8s_node"
	CapabilityK8sWorkload                  = "k8s_workload"
	CapabilityK8sPod                       = "k8s_pod"
	CapabilityK8sIngress                   = "k8s_ingress"
	CapabilityK8sService                   = "k8s_service"
	CapabilityK8sConfigMap                 = "k8s_configmap"
	CapabilityK8sSecret                    = "k8s_secret"
	CapabilityK8sPV                        = "k8s_pv"
	CapabilityK8sPVC                       = "k8s_pvc"
	CapabilityK8sJob                       = "k8s_job"
	CapabilityK8sCronJob                   = "k8s_cronjob"
	CapabilityAgentMachine                 = "agent_machine"
	CapabilitySSHCommand                   = "ssh_command"
	CapabilityWebShell                     = "web_shell"
)

// Capability 是一个能力及其包含的租户权限。
// Permissions 的顺序与 Catalog 一致，供套餐勾选和角色授权页按能力展示。
type Capability struct {
	Code        string
	Module      string
	Name        string
	Description string
	Permissions []string
}

// capabilitySpecs 是能力目录的唯一来源，顺序即 Capabilities() 与设计文档的展示顺序。
// 模块码复用 featureOrder 里的 Feature 常量；权限码复用 perm.go 里的常量。
var capabilitySpecs = []Capability{
	{
		Code: CapabilityAlertRealtime, Module: FeatureAlert, Name: "实时告警",
		Description: "查看实时告警，处理与干预告警事件。",
		Permissions: []string{AlertRealtimeRead, AlertRealtimeIntervene},
	},
	{
		Code: CapabilityAlertHistory, Module: FeatureAlert, Name: "历史告警",
		Description: "查看历史告警与恢复分析。",
		Permissions: []string{AlertHistoryRead},
	},
	{
		Code: CapabilityAlertPage, Module: FeatureAlert, Name: "告警页面",
		Description: "配置实时告警的标签页面。",
		Permissions: []string{AlertPageRead, AlertPageWrite, AlertPageDelete},
	},
	{
		Code: CapabilityAlertRule, Module: FeatureAlertRule, Name: "告警规则",
		Description: "管理告警规则、规则组与告警级别。",
		Permissions: []string{AlertRuleRead, AlertRuleWrite, AlertRuleDelete},
	},
	{
		Code: CapabilityAlertDictionary, Module: FeatureAlertRule, Name: "告警字典",
		Description: "维护告警级别与数据源级别字典。",
		Permissions: []string{AlertDictionaryRead, AlertDictionaryWrite, AlertDictionaryDelete},
	},
	{
		Code: CapabilityProbeHTTP, Module: FeatureProbe, Name: "HTTP 探测",
		Description: "管理 HTTP 探测目标组、探测地址与立即探测。",
		Permissions: []string{ProbeRead, ProbeWrite, ProbeDelete},
	},
	{
		Code: CapabilityDatasourcePrometheus, Module: FeatureDatasource, Name: "Prometheus 指标",
		Description: "接入 Prometheus 兼容的 PromQL 指标数据源。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityDatasourceVictoriaMetrics, Module: FeatureDatasource, Name: "VictoriaMetrics 指标",
		Description: "接入 VictoriaMetrics 指标数据源。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityDatasourceElasticsearch, Module: FeatureDatasource, Name: "Elasticsearch 日志",
		Description: "接入 Elasticsearch 日志数据源并按时间桶评测。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityDatasourceLoki, Module: FeatureDatasource, Name: "Loki 日志",
		Description: "接入 Loki 日志数据源并执行 LogQL 数值型评测。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityDatasourceVictoriaLogs, Module: FeatureDatasource, Name: "VictoriaLogs 日志",
		Description: "接入 VictoriaLogs 数据源并执行 LogsQL 数值型评测。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityDatasourceJaeger, Module: FeatureDatasource, Name: "Jaeger 链路",
		Description: "接入 Jaeger 链路数据源。",
		Permissions: []string{DatasourceRead, DatasourceWrite, DatasourceDelete},
	},
	{
		Code: CapabilityNotifyChannelSMTP, Module: FeatureNotification, Name: "邮件渠道",
		Description: "配置 SMTP 邮件通知渠道。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelWebhookOther, Module: FeatureNotification, Name: "自定义 Webhook",
		Description: "配置自定义 Webhook 通知渠道，服务端以 JSON 发送并附带签名。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelWebhookDingTalk, Module: FeatureNotification, Name: "钉钉 Webhook",
		Description: "配置钉钉自定义机器人 Webhook 通知渠道。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelWebhookWeChat, Module: FeatureNotification, Name: "企业微信 Webhook",
		Description: "配置企业微信群机器人 Webhook 通知渠道。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelWebhookFeishu, Module: FeatureNotification, Name: "飞书 Webhook",
		Description: "配置飞书自定义机器人 Webhook 通知渠道。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelSMSAliCloud, Module: FeatureNotification, Name: "阿里云短信",
		Description: "配置阿里云短信通知渠道。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyChannelSMSTencent, Module: FeatureNotification, Name: "腾讯云短信",
		Description: "开通腾讯云短信渠道。当前版本只上架能力，暂不支持配置与发送。",
		Permissions: []string{NotificationChannelRead, NotificationChannelWrite, NotificationChannelDelete},
	},
	{
		Code: CapabilityNotifyTemplate, Module: FeatureNotification, Name: "消息模板",
		Description: "维护通知消息模板。",
		Permissions: []string{NotificationTemplateRead, NotificationTemplateWrite, NotificationTemplateDelete},
	},
	{
		Code: CapabilityNotifyGroup, Module: FeatureNotification, Name: "通知组",
		Description: "组合渠道、模板与接收人。",
		Permissions: []string{NotificationGroupRead, NotificationGroupWrite, NotificationGroupDelete},
	},
	{
		Code: CapabilityNotifySubscription, Module: FeatureNotification, Name: "告警订阅",
		Description: "按标签与级别订阅告警并绑定通知组。",
		Permissions: []string{NotificationSubscriptionRead, NotificationSubscriptionWrite, NotificationSubscriptionDelete},
	},
	{
		Code: CapabilityNotifyOnCall, Module: FeatureNotification, Name: "OnCall 值班",
		Description: "按时段把告警路由给不同通知组。",
		Permissions: []string{NotificationOncallRead, NotificationOncallWrite, NotificationOncallDelete},
	},
	{
		Code: CapabilityMessageLog, Module: FeatureNotification, Name: "消息日志",
		Description: "查看通知投递记录与失败原因。",
		Permissions: []string{MessageLogRead, MessageLogWrite, MessageLogDelete},
	},
	{
		Code: CapabilityMember, Module: FeatureMember, Name: "成员",
		Description: "邀请、移除成员并分配角色。",
		Permissions: []string{MemberRead, MemberWrite, MemberDelete},
	},
	{
		Code: CapabilityRole, Module: FeatureRole, Name: "角色",
		Description: "管理租户角色和角色权限。",
		Permissions: []string{RoleRead, RoleWrite, RoleDelete},
	},
	{
		Code: CapabilityClusterAccess, Module: FeatureCluster, Name: "集群接入",
		Description: "接入 Kubernetes 集群，查看详情、测试连通性与修改 kubeconfig。",
		Permissions: []string{ClusterRead, ClusterWrite, ClusterDelete},
	},
	{
		Code: CapabilityK8sNode, Module: FeatureCluster, Name: "节点",
		Description: "查看 Kubernetes 节点与节点运维。",
		Permissions: []string{K8sNodeRead, K8sNodeWrite},
	},
	{
		Code: CapabilityK8sWorkload, Module: FeatureCluster, Name: "工作负载",
		Description: "查看与运维 Deployment、StatefulSet、DaemonSet。",
		Permissions: []string{K8sWorkloadRead, K8sWorkloadWrite},
	},
	{
		Code: CapabilityK8sPod, Module: FeatureCluster, Name: "Pod",
		Description: "查看 Pod、删除 Pod 与读取容器日志。",
		Permissions: []string{K8sPodRead, K8sPodWrite, K8sPodLogs},
	},
	{
		Code: CapabilityK8sIngress, Module: FeatureCluster, Name: "Ingress",
		Description: "查看 Kubernetes Ingress。",
		Permissions: []string{K8sIngressRead},
	},
	{
		Code: CapabilityK8sService, Module: FeatureCluster, Name: "Service",
		Description: "查看 Kubernetes Service。",
		Permissions: []string{K8sServiceRead},
	},
	{
		Code: CapabilityK8sConfigMap, Module: FeatureCluster, Name: "ConfigMap",
		Description: "查看 Kubernetes ConfigMap。",
		Permissions: []string{K8sConfigMapRead},
	},
	{
		Code: CapabilityK8sSecret, Module: FeatureCluster, Name: "Secret",
		Description: "查看 Kubernetes Secret。",
		Permissions: []string{K8sSecretRead},
	},
	{
		Code: CapabilityK8sPV, Module: FeatureCluster, Name: "PV",
		Description: "查看 Kubernetes PersistentVolume。",
		Permissions: []string{K8sPersistentVolumeRead},
	},
	{
		Code: CapabilityK8sPVC, Module: FeatureCluster, Name: "PVC",
		Description: "查看 Kubernetes PersistentVolumeClaim。",
		Permissions: []string{K8sPersistentVolumeClaimRead},
	},
	{
		Code: CapabilityK8sJob, Module: FeatureCluster, Name: "Job",
		Description: "查看与删除 Kubernetes Job。",
		Permissions: []string{K8sJobRead, K8sJobWrite},
	},
	{
		Code: CapabilityK8sCronJob, Module: FeatureCluster, Name: "CronJob",
		Description: "查看与运维 Kubernetes CronJob。",
		Permissions: []string{K8sCronJobRead, K8sCronJobWrite},
	},
	{
		Code: CapabilityAgentMachine, Module: FeatureAgent, Name: "机器",
		Description: "管理边缘机器和 Agent 状态。",
		Permissions: []string{AgentRead, AgentWrite, AgentDelete},
	},
	{
		Code: CapabilitySSHCommand, Module: FeatureSSH, Name: "SSH 命令",
		Description: "维护 SSH 命令并查看执行审计。",
		Permissions: []string{SSHRead, SSHWrite, SSHDelete},
	},
	{
		Code: CapabilityWebShell, Module: FeatureSSH, Name: "Web Shell",
		Description: "在浏览器里对机器执行 SSH 命令。",
		Permissions: []string{SSHExec},
	},
}

// CapabilityMember、CapabilityRole 等能力码与模块码同名，是刻意为之：
// 模块下只有一个能力时，两者语义完全一致，前端可以直接复用编码。
// 数据源、探测和通知渠道拆成多个能力后，模块下不再只有一个能力，模块由能力反推。
var (
	capabilityIndex = func() map[string]Capability {
		index := make(map[string]Capability, len(capabilitySpecs))
		for _, item := range capabilitySpecs {
			index[item.Code] = item
		}
		return index
	}()
)

// DatasourceCapabilities 按目录顺序返回全部数据源驱动能力。
func DatasourceCapabilities() []string {
	return CapabilityCodesOfModule(FeatureDatasource)
}

// Capabilities 按固定顺序返回全部能力，供套餐勾选和前端分组。
func Capabilities() []Capability {
	items := make([]Capability, 0, len(capabilitySpecs))
	for _, item := range capabilitySpecs {
		permissions := make([]string, len(item.Permissions))
		copy(permissions, item.Permissions)
		items = append(items, Capability{
			Code:        item.Code,
			Module:      item.Module,
			Name:        item.Name,
			Description: item.Description,
			Permissions: permissions,
		})
	}
	return items
}

// CapabilityOf 返回能力详情；能力码不存在时第二个返回值为 false。
func CapabilityOf(code string) (Capability, bool) {
	item, ok := capabilityIndex[code]
	if !ok {
		return Capability{}, false
	}
	permissions := make([]string, len(item.Permissions))
	copy(permissions, item.Permissions)
	item.Permissions = permissions
	return item, true
}

// IsCapability 判断能力编码是否在目录里。
func IsCapability(code string) bool {
	_, ok := capabilityIndex[code]
	return ok
}

// CapabilitiesOfModule 按目录顺序返回某个模块下的全部能力。
func CapabilitiesOfModule(module string) []Capability {
	items := make([]Capability, 0, 1)
	for _, code := range CapabilityCodesOfModule(module) {
		item, _ := CapabilityOf(code)
		items = append(items, item)
	}
	return items
}

// CapabilityCodesOfModule 按目录顺序返回某个模块下的全部能力编码。
// seed 把旧的模块级套餐明细展开成能力时用它。
func CapabilityCodesOfModule(module string) []string {
	codes := make([]string, 0, 1)
	for _, item := range capabilitySpecs {
		if item.Module == module {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// legacyCapabilityExpansion 是历史版本的能力码到当前能力码的展开表。
// M21 把模块级能力拆成类型，M24 又把通知渠道类型拆成 Webhook 应用和短信提供商；
// 升级迁移、旧套餐快照和旧客户端提交都要按这张表归一化，不能让存量租户丢能力。
// 这些编码在当前目录里已经不存在，只能作为迁移/兼容输入。
var legacyCapabilityExpansion = map[string][]string{
	// 旧的模块级「数据源」覆盖全部驱动；它同时也是当前模块码，展开结果与
	// CapabilityCodesOfModule(FeatureDatasource) 一致。
	"datasource": {
		CapabilityDatasourcePrometheus,
		CapabilityDatasourceVictoriaMetrics,
		CapabilityDatasourceElasticsearch,
		CapabilityDatasourceLoki,
		CapabilityDatasourceVictoriaLogs,
		CapabilityDatasourceJaeger,
	},
	"probe_target": {CapabilityProbeHTTP},
	"notify_channel": {
		CapabilityNotifyChannelSMTP,
		CapabilityNotifyChannelWebhookOther,
		CapabilityNotifyChannelWebhookDingTalk,
		CapabilityNotifyChannelWebhookWeChat,
		CapabilityNotifyChannelWebhookFeishu,
		CapabilityNotifyChannelSMSAliCloud,
		CapabilityNotifyChannelSMSTencent,
	},
	"notify_channel_webhook": {
		CapabilityNotifyChannelWebhookOther,
		CapabilityNotifyChannelWebhookDingTalk,
		CapabilityNotifyChannelWebhookWeChat,
		CapabilityNotifyChannelWebhookFeishu,
	},
	"notify_channel_sms": {
		CapabilityNotifyChannelSMSAliCloud,
		CapabilityNotifyChannelSMSTencent,
	},
}

// IsLegacyCapability 判断编码是不是历史能力码（能展开成当前能力）。
// 它和 IsCapability 互斥：历史编码不在当前能力目录里。
func IsLegacyCapability(code string) bool {
	_, ok := legacyCapabilityExpansion[code]
	return ok
}

// ExpandLegacyCapability 返回历史能力码对应的当前能力，按目录顺序去重；
// 非历史编码返回 nil。
func ExpandLegacyCapability(code string) []string {
	codes, ok := legacyCapabilityExpansion[code]
	if !ok {
		return nil
	}
	return EffectiveCapabilities(nil, codes)
}

// NormalizeCapabilities 把一组编码归一化成当前能力目录里的编码，按目录顺序去重：
//   - 目录里的能力码保留；
//   - 模块码展开成该模块的全部能力；
//   - 历史编码按 legacyCapabilityExpansion 展开；
//   - 完全未知的编码丢弃（需要严格校验的入口先用 IsCapability / IsLegacyCapability 判断）。
func NormalizeCapabilities(codes []string) []string {
	expanded := make([]string, 0, len(codes))
	for _, code := range codes {
		switch {
		case code == "":
			continue
		case IsCapability(code):
			expanded = append(expanded, code)
		case IsFeature(code):
			expanded = append(expanded, CapabilityCodesOfModule(code)...)
		default:
			expanded = append(expanded, legacyCapabilityExpansion[code]...)
		}
	}
	return EffectiveCapabilities(nil, expanded)
}

// BaseCapabilities 返回基础套餐包含的能力：前七个模块的全部能力，共 26 项。
// 数据源、探测和通知渠道按类型/应用展开后全部包含在基础套餐里，保持升级前
// 「基础套餐可用全部数据源与通知渠道类型」的行为；Kubernetes、机器和 SSH
// 能力由超级管理员另行开通。
func BaseCapabilities() []string {
	modules := make(map[string]struct{}, len(BaseFeatures()))
	for _, code := range BaseFeatures() {
		modules[code] = struct{}{}
	}
	codes := make([]string, 0, 26)
	for _, item := range capabilitySpecs {
		if _, ok := modules[item.Module]; ok {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// PermissionCapabilities 返回包含该权限的能力编码，按能力目录顺序。
// 角色授权页用它把接口权限归到某个菜单/能力下；多个能力共享权限时取第一个。
func PermissionCapabilities(permission string) []string {
	out := make([]string, 0, 2)
	for _, item := range capabilitySpecs {
		if slices.Contains(item.Permissions, permission) {
			out = append(out, item.Code)
		}
	}
	return out
}

// CapabilityPermissions 返回一组能力的权限并集，按能力目录顺序去重。
// 同一个权限可能属于多个能力（例如通知渠道的七个能力共享 notification_channel:*），
// 这里只保留一份。
func CapabilityPermissions(capabilities []string) []string {
	seenCapability := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		seenCapability[code] = struct{}{}
	}
	seenPermission := make(map[string]struct{}, len(seenCapability)*3)
	names := make([]string, 0, len(seenCapability)*3)
	for _, item := range capabilitySpecs {
		if _, ok := seenCapability[item.Code]; !ok {
			continue
		}
		for _, name := range item.Permissions {
			if _, ok := seenPermission[name]; ok {
				continue
			}
			seenPermission[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

// EffectiveCapabilities 计算套餐能力与租户额外能力的并集，按能力目录顺序去重。
// 旧数据里不在目录内的编码被忽略；租户差异只加不减，由仓储在写入时保证。
func EffectiveCapabilities(packageCapabilities, extraCapabilities []string) []string {
	allowed := make(map[string]struct{}, len(packageCapabilities)+len(extraCapabilities))
	for _, code := range packageCapabilities {
		allowed[code] = struct{}{}
	}
	for _, code := range extraCapabilities {
		allowed[code] = struct{}{}
	}
	codes := make([]string, 0, len(allowed))
	for _, item := range capabilitySpecs {
		if _, ok := allowed[item.Code]; ok {
			codes = append(codes, item.Code)
		}
	}
	return codes
}

// ModulesOfCapabilities 把能力集合反推成模块集合，按模块展示顺序返回。
// 有效模块用于 access 的模块级功能检查：模块里还有一个能力开通，模块就还在。
func ModulesOfCapabilities(capabilities []string) []string {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		item, ok := CapabilityOf(code)
		if !ok {
			continue
		}
		enabled[item.Module] = struct{}{}
	}
	modules := make([]string, 0, len(enabled))
	for _, code := range featureOrder {
		if _, ok := enabled[code]; ok {
			modules = append(modules, code)
		}
	}
	return modules
}

// datasourceCapabilityDrivers 是数据源驱动与能力编码的唯一映射，顺序与
// datasource.driver 枚举和告警规则 driver 的展示顺序一致。
var datasourceCapabilityDrivers = []struct {
	driver     string
	capability string
}{
	{"METRICS_PROMETHEUS", CapabilityDatasourcePrometheus},
	{"METRICS_VICTORIA_METRICS", CapabilityDatasourceVictoriaMetrics},
	{"LOGS_ELASTICSEARCH", CapabilityDatasourceElasticsearch},
	{"LOGS_LOKI", CapabilityDatasourceLoki},
	{"LOGS_VLOGS", CapabilityDatasourceVictoriaLogs},
	{"TRACE_JAEGER", CapabilityDatasourceJaeger},
}

// DatasourceDrivers 按固定顺序返回全部数据源驱动编码。
func DatasourceDrivers() []string {
	out := make([]string, 0, len(datasourceCapabilityDrivers))
	for _, item := range datasourceCapabilityDrivers {
		out = append(out, item.driver)
	}
	return out
}

// DatasourceCapabilityOf 返回数据源驱动对应的能力编码；未知驱动返回空字符串。
// 数据源创建、数据源列表过滤和告警规则的类型检查都从这里取能力。
func DatasourceCapabilityOf(driver string) string {
	for _, item := range datasourceCapabilityDrivers {
		if item.driver == driver {
			return item.capability
		}
	}
	return ""
}

// DatasourceDriverOf 返回能力编码对应的数据源驱动；不是数据源能力时返回空。
func DatasourceDriverOf(capability string) (string, bool) {
	for _, item := range datasourceCapabilityDrivers {
		if item.capability == capability {
			return item.driver, true
		}
	}
	return "", false
}

// DatasourceDriversOf 返回已开通的数据源驱动，按目录顺序；空结果表示一个类型都没开通。
func DatasourceDriversOf(capabilities []string) []string {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	drivers := make([]string, 0, len(datasourceCapabilityDrivers))
	for _, item := range datasourceCapabilityDrivers {
		if _, ok := enabled[item.capability]; ok {
			drivers = append(drivers, item.driver)
		}
	}
	return drivers
}

// probeCapabilityDrivers 是探测驱动与能力编码的映射。一期只有 HTTP；
// 新增 TCP / ICMP 时在这里追加一行并同时实现执行器，不能先卖能力。
var probeCapabilityDrivers = []struct {
	driver     string
	capability string
}{
	{"PROBE_HTTP", CapabilityProbeHTTP},
}

// ProbeDrivers 按固定顺序返回全部探测驱动编码。
func ProbeDrivers() []string {
	out := make([]string, 0, len(probeCapabilityDrivers))
	for _, item := range probeCapabilityDrivers {
		out = append(out, item.driver)
	}
	return out
}

// ProbeCapabilityOf 返回探测驱动对应的能力编码；未知驱动返回空字符串。
func ProbeCapabilityOf(driver string) string {
	for _, item := range probeCapabilityDrivers {
		if item.driver == driver {
			return item.capability
		}
	}
	return ""
}

// ProbeDriversOf 返回已开通的探测驱动，按目录顺序。
func ProbeDriversOf(capabilities []string) []string {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	drivers := make([]string, 0, len(probeCapabilityDrivers))
	for _, item := range probeCapabilityDrivers {
		if _, ok := enabled[item.capability]; ok {
			drivers = append(drivers, item.driver)
		}
	}
	return drivers
}

// ChannelScope 是一个可授权的通知渠道「渠道类型 + 应用」组合。
// Webhook 按厂商、短信按提供商拆分能力后，渠道的读写与列表都要按组合判断，
// 不能只看渠道类型。
type ChannelScope struct {
	Type string
	App  string
}

// TemplateScope 是模板的「模板类型 + 应用」组合。模板是渠道的正文，跟着对应
// 渠道能力授权；腾讯云短信没有模板类型，不在映射里。
type TemplateScope struct {
	Type string
	App  string
}

// channelTypeOrder 是渠道类型枚举顺序，列表与下拉都按它展示。
var channelTypeOrder = []string{"SMTP", "WEBHOOK", "SMS"}

// templateTypeOrder 是模板类型枚举顺序。
var templateTypeOrder = []string{"EMAIL", "WEBHOOK", "SMS_ALICLOUD"}

// channelCapabilityScopes 是通知渠道 (类型, 应用) 与能力编码的映射，顺序与渠道
// 类型、应用枚举一致。腾讯云短信只上架能力、暂不支持配置渠道，所以没有渠道组合。
var channelCapabilityScopes = []struct {
	scope      ChannelScope
	capability string
}{
	{ChannelScope{Type: "SMTP", App: "OTHER"}, CapabilityNotifyChannelSMTP},
	{ChannelScope{Type: "WEBHOOK", App: "OTHER"}, CapabilityNotifyChannelWebhookOther},
	{ChannelScope{Type: "WEBHOOK", App: "DINGTALK"}, CapabilityNotifyChannelWebhookDingTalk},
	{ChannelScope{Type: "WEBHOOK", App: "WECHAT"}, CapabilityNotifyChannelWebhookWeChat},
	{ChannelScope{Type: "WEBHOOK", App: "FEISHU"}, CapabilityNotifyChannelWebhookFeishu},
	{ChannelScope{Type: "SMS", App: "SMS_ALICLOUD"}, CapabilityNotifyChannelSMSAliCloud},
}

// templateCapabilityScopes 是模板 (类型, 应用) 与渠道能力编码的映射。
// EMAIL 需要 SMTP，WEBHOOK 需要对应 Webhook 应用能力，SMS_ALICLOUD 需要
// 阿里云短信能力；腾讯云短信没有模板类型。
var templateCapabilityScopes = []struct {
	scope      TemplateScope
	capability string
}{
	{TemplateScope{Type: "EMAIL", App: "OTHER"}, CapabilityNotifyChannelSMTP},
	{TemplateScope{Type: "WEBHOOK", App: "OTHER"}, CapabilityNotifyChannelWebhookOther},
	{TemplateScope{Type: "WEBHOOK", App: "DINGTALK"}, CapabilityNotifyChannelWebhookDingTalk},
	{TemplateScope{Type: "WEBHOOK", App: "WECHAT"}, CapabilityNotifyChannelWebhookWeChat},
	{TemplateScope{Type: "WEBHOOK", App: "FEISHU"}, CapabilityNotifyChannelWebhookFeishu},
	{TemplateScope{Type: "SMS_ALICLOUD", App: "OTHER"}, CapabilityNotifyChannelSMSAliCloud},
}

// ChannelTypes 按固定顺序返回全部通知渠道类型。
func ChannelTypes() []string {
	out := make([]string, len(channelTypeOrder))
	copy(out, channelTypeOrder)
	return out
}

// ChannelCapabilityOf 返回通知渠道 (类型, 应用) 对应的能力编码；未知组合返回空字符串。
func ChannelCapabilityOf(channelType, app string) string {
	for _, item := range channelCapabilityScopes {
		if item.scope.Type == channelType && item.scope.App == app {
			return item.capability
		}
	}
	return ""
}

// ChannelScopesOf 返回已开通的通知渠道 (类型, 应用) 组合，按目录顺序。
// 列表、选择器和通知组过滤都按这个组合集合查询，未授权的应用不会出现。
func ChannelScopesOf(capabilities []string) []ChannelScope {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	scopes := make([]ChannelScope, 0, len(channelCapabilityScopes))
	for _, item := range channelCapabilityScopes {
		if _, ok := enabled[item.capability]; ok {
			scopes = append(scopes, item.scope)
		}
	}
	return scopes
}

// ChannelTypesOf 返回已开通的通知渠道类型，按目录顺序。
func ChannelTypesOf(capabilities []string) []string {
	authorized := make(map[string]struct{})
	for _, scope := range ChannelScopesOf(capabilities) {
		authorized[scope.Type] = struct{}{}
	}
	types := make([]string, 0, len(channelTypeOrder))
	for _, channelType := range channelTypeOrder {
		if _, ok := authorized[channelType]; ok {
			types = append(types, channelType)
		}
	}
	return types
}

// ChannelAppsOf 返回某类渠道已开通的应用，按应用枚举顺序。
// SMTP 固定返回 OTHER；短信只返回实现了发送能力的阿里云。
func ChannelAppsOf(channelType string, capabilities []string) []string {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	apps := make([]string, 0, len(channelCapabilityScopes))
	seen := make(map[string]struct{})
	for _, item := range channelCapabilityScopes {
		if item.scope.Type != channelType {
			continue
		}
		if _, ok := enabled[item.capability]; !ok {
			continue
		}
		if _, ok := seen[item.scope.App]; ok {
			continue
		}
		seen[item.scope.App] = struct{}{}
		apps = append(apps, item.scope.App)
	}
	return apps
}

// TemplateCapabilityOf 返回模板 (类型, 应用) 对应的渠道能力编码；
// 未知组合返回空字符串。
func TemplateCapabilityOf(templateType, app string) string {
	for _, item := range templateCapabilityScopes {
		if item.scope.Type == templateType && item.scope.App == app {
			return item.capability
		}
	}
	return ""
}

// TemplateScopesOf 返回已开通的模板 (类型, 应用) 组合，按目录顺序。
func TemplateScopesOf(capabilities []string) []TemplateScope {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	scopes := make([]TemplateScope, 0, len(templateCapabilityScopes))
	for _, item := range templateCapabilityScopes {
		if _, ok := enabled[item.capability]; ok {
			scopes = append(scopes, item.scope)
		}
	}
	return scopes
}

// TemplateTypesOf 返回已开通的模板类型，按模板类型枚举顺序。
func TemplateTypesOf(capabilities []string) []string {
	authorized := make(map[string]struct{})
	for _, scope := range TemplateScopesOf(capabilities) {
		authorized[scope.Type] = struct{}{}
	}
	types := make([]string, 0, len(templateTypeOrder))
	for _, templateType := range templateTypeOrder {
		if _, ok := authorized[templateType]; ok {
			types = append(types, templateType)
		}
	}
	return types
}

// TemplateAppsOf 返回某类模板已开通的应用，按应用枚举顺序。
// EMAIL 和 SMS_ALICLOUD 固定返回 OTHER；Webhook 模板按应用能力展开。
func TemplateAppsOf(templateType string, capabilities []string) []string {
	enabled := make(map[string]struct{}, len(capabilities))
	for _, code := range capabilities {
		enabled[code] = struct{}{}
	}
	apps := make([]string, 0, len(templateCapabilityScopes))
	seen := make(map[string]struct{})
	for _, item := range templateCapabilityScopes {
		if item.scope.Type != templateType {
			continue
		}
		if _, ok := enabled[item.capability]; !ok {
			continue
		}
		if _, ok := seen[item.scope.App]; ok {
			continue
		}
		seen[item.scope.App] = struct{}{}
		apps = append(apps, item.scope.App)
	}
	return apps
}
