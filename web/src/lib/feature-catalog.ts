// 能力目录（套餐 / 租户功能授权 / 菜单可见性）。
//
// 能力 code 与后端 perm.Features() 的 30 项定义保持一致；GET /v1/system/features
// 返回的是同一份目录（code/module/name/description/permissions）。页面用这里的
// 双语标签，不直接用接口里的 name/description，这样切到 en 时中文名也不会消失，
// 后端没部署新接口时也能先用本地目录渲染勾选框。
//
// 目录分两级：MODULES 是所属菜单模块（10 个），CAPABILITIES 是具体能力（30 个）。
// 新增 / 删除能力时同步：后端 perm.Features()、本文件、i18n 里的菜单文案。

export type LocalizedLabel = { zh: string; en: string }

export type ModuleDefinition = {
  code: string
  label: LocalizedLabel
}

export type CapabilityDefinition = {
  code: string
  /** 所属菜单模块 code，对应 MODULES 的一项。 */
  module: string
  label: LocalizedLabel
  description: LocalizedLabel
}

/** 模块顺序：告警 → 探测 → 通知 → 租户 → Kubernetes → 边缘。 */
export const MODULES: ModuleDefinition[] = [
  { code: "alert", label: { zh: "告警", en: "Alerts" } },
  { code: "alert_rule", label: { zh: "告警规则", en: "Alert rules" } },
  { code: "probe", label: { zh: "探测目标", en: "Probe targets" } },
  { code: "datasource", label: { zh: "数据源", en: "Datasources" } },
  { code: "notification", label: { zh: "通知", en: "Notifications" } },
  { code: "member", label: { zh: "成员", en: "Members" } },
  { code: "role", label: { zh: "角色", en: "Roles" } },
  { code: "cluster", label: { zh: "Kubernetes", en: "Kubernetes" } },
  { code: "agent", label: { zh: "机器", en: "Machines" } },
  { code: "ssh", label: { zh: "SSH 与 Web Shell", en: "SSH & web shell" } },
]

/**
 * 能力目录：顺序与后端 perm.Features() 一致。
 * 基础套餐 = 前 7 个模块（alert / alert_rule / probe / datasource /
 * notification / member / role）的全部能力，共 15 项。
 */
export const CAPABILITIES: CapabilityDefinition[] = [
  {
    code: "alert_realtime",
    module: "alert",
    label: { zh: "实时告警", en: "Realtime alerts" },
    description: {
      zh: "正在触发的告警列表、详情与干预",
      en: "Firing alerts with live details and intervention",
    },
  },
  {
    code: "alert_history",
    module: "alert",
    label: { zh: "历史告警", en: "Historical alerts" },
    description: {
      zh: "历史事件台账与统计分析",
      en: "Historical event ledger and analysis",
    },
  },
  {
    code: "alert_page",
    module: "alert",
    label: { zh: "告警页面", en: "Alert pages" },
    description: {
      zh: "自定义告警展示页面",
      en: "Custom alert display pages",
    },
  },
  {
    code: "alert_rule",
    module: "alert_rule",
    label: { zh: "告警规则", en: "Alert rules" },
    description: {
      zh: "告警规则、评测与生效状态",
      en: "Alert rules, evaluation, and activation",
    },
  },
  {
    code: "alert_dictionary",
    module: "alert_rule",
    label: { zh: "告警字典", en: "Alert dictionaries" },
    description: {
      zh: "告警等级与字典项维护",
      en: "Alert levels and dictionary entries",
    },
  },
  {
    code: "probe_http",
    module: "probe",
    label: { zh: "HTTP 探测", en: "HTTP probes" },
    description: {
      zh: "HTTP 探测目标组与探测地址",
      en: "HTTP probe target groups and addresses",
    },
  },
  {
    code: "datasource_prometheus",
    module: "datasource",
    label: { zh: "Prometheus 指标", en: "Prometheus metrics" },
    description: {
      zh: "Prometheus 兼容的 PromQL 指标数据源",
      en: "Prometheus-compatible PromQL datasources",
    },
  },
  {
    code: "datasource_victoria_metrics",
    module: "datasource",
    label: { zh: "VictoriaMetrics 指标", en: "VictoriaMetrics metrics" },
    description: {
      zh: "VictoriaMetrics 指标数据源",
      en: "VictoriaMetrics metrics datasources",
    },
  },
  {
    code: "datasource_elasticsearch",
    module: "datasource",
    label: { zh: "Elasticsearch 日志", en: "Elasticsearch logs" },
    description: {
      zh: "Elasticsearch 日志数据源与时间桶评测",
      en: "Elasticsearch log datasources and time-bucket evaluation",
    },
  },
  {
    code: "datasource_loki",
    module: "datasource",
    label: { zh: "Loki 日志", en: "Loki logs" },
    description: {
      zh: "Loki 日志数据源与 LogQL 评测",
      en: "Loki log datasources and LogQL evaluation",
    },
  },
  {
    code: "datasource_vlogs",
    module: "datasource",
    label: { zh: "VictoriaLogs 日志", en: "VictoriaLogs logs" },
    description: {
      zh: "VictoriaLogs 日志数据源与 LogsQL 评测",
      en: "VictoriaLogs datasources and LogsQL evaluation",
    },
  },
  {
    code: "datasource_jaeger",
    module: "datasource",
    label: { zh: "Jaeger 链路", en: "Jaeger traces" },
    description: {
      zh: "Jaeger 链路数据源",
      en: "Jaeger trace datasources",
    },
  },
  {
    code: "notify_channel_smtp",
    module: "notification",
    label: { zh: "邮件渠道", en: "Email channels" },
    description: {
      zh: "SMTP 邮件通知渠道",
      en: "SMTP email notification channels",
    },
  },
  {
    code: "notify_channel_webhook",
    module: "notification",
    label: { zh: "Webhook 渠道", en: "Webhook channels" },
    description: {
      zh: "Webhook 通知渠道（含钉钉、企业微信、飞书）",
      en: "Webhook channels including DingTalk, WeCom, and Feishu",
    },
  },
  {
    code: "notify_channel_sms",
    module: "notification",
    label: { zh: "短信渠道", en: "SMS channels" },
    description: {
      zh: "短信通知渠道",
      en: "SMS notification channels",
    },
  },
  {
    code: "notify_template",
    module: "notification",
    label: { zh: "消息模板", en: "Message templates" },
    description: {
      zh: "通知内容模板与变量",
      en: "Notification content templates and variables",
    },
  },
  {
    code: "notify_group",
    module: "notification",
    label: { zh: "通知组", en: "Notification groups" },
    description: {
      zh: "接收通知的成员分组",
      en: "Member groups that receive notifications",
    },
  },
  {
    code: "notify_subscription",
    module: "notification",
    label: { zh: "告警订阅", en: "Alert subscriptions" },
    description: {
      zh: "规则与通知组的订阅关系",
      en: "Subscriptions between rules and notification groups",
    },
  },
  {
    code: "notify_oncall",
    module: "notification",
    label: { zh: "OnCall 值班", en: "On-call schedules" },
    description: {
      zh: "值班表、轮换与替班",
      en: "On-call rosters, rotations, and substitutions",
    },
  },
  {
    code: "message_log",
    module: "notification",
    label: { zh: "消息日志", en: "Message logs" },
    description: {
      zh: "通知任务的投递记录",
      en: "Delivery records for notification tasks",
    },
  },
  {
    code: "member",
    module: "member",
    label: { zh: "成员", en: "Members" },
    description: {
      zh: "租户成员管理与邀请",
      en: "Tenant membership and invitations",
    },
  },
  {
    code: "role",
    module: "role",
    label: { zh: "角色", en: "Roles" },
    description: {
      zh: "角色与角色权限分配",
      en: "Roles and role permission assignment",
    },
  },
  {
    code: "cluster_access",
    module: "cluster",
    label: { zh: "集群接入", en: "Cluster access" },
    description: {
      zh: "Kubernetes 集群接入、凭据与连通性",
      en: "Kubernetes cluster registration, credentials, and connectivity",
    },
  },
  {
    code: "k8s_node",
    module: "cluster",
    label: { zh: "节点", en: "Nodes" },
    description: { zh: "集群节点与资源使用", en: "Cluster nodes and usage" },
  },
  {
    code: "k8s_workload",
    module: "cluster",
    label: { zh: "工作负载", en: "Workloads" },
    description: {
      zh: "Deployment / StatefulSet / DaemonSet",
      en: "Deployment, StatefulSet, and DaemonSet",
    },
  },
  {
    code: "k8s_pod",
    module: "cluster",
    label: { zh: "Pod", en: "Pods" },
    description: {
      zh: "Pod 列表、日志与终端",
      en: "Pod list, logs, and terminal",
    },
  },
  {
    code: "k8s_ingress",
    module: "cluster",
    label: { zh: "Ingress", en: "Ingresses" },
    description: { zh: "路由入口与后端服务", en: "Routing rules and backends" },
  },
  {
    code: "k8s_service",
    module: "cluster",
    label: { zh: "Service", en: "Services" },
    description: { zh: "服务与端口映射", en: "Services and port mappings" },
  },
  {
    code: "k8s_configmap",
    module: "cluster",
    label: { zh: "ConfigMap", en: "ConfigMaps" },
    description: { zh: "配置项与文件", en: "Configuration entries and files" },
  },
  {
    code: "k8s_secret",
    module: "cluster",
    label: { zh: "Secret", en: "Secrets" },
    description: { zh: "密钥与凭据", en: "Secrets and credentials" },
  },
  {
    code: "k8s_pv",
    module: "cluster",
    label: { zh: "PV", en: "Persistent volumes" },
    description: {
      zh: "集群级持久卷",
      en: "Cluster-level persistent volumes",
    },
  },
  {
    code: "k8s_pvc",
    module: "cluster",
    label: { zh: "PVC", en: "Persistent volume claims" },
    description: {
      zh: "命名空间存储申请",
      en: "Namespaced persistent volume claims",
    },
  },
  {
    code: "k8s_job",
    module: "cluster",
    label: { zh: "Job", en: "Jobs" },
    description: { zh: "一次性任务", en: "One-off jobs" },
  },
  {
    code: "k8s_cronjob",
    module: "cluster",
    label: { zh: "CronJob", en: "CronJobs" },
    description: { zh: "定时任务", en: "Scheduled jobs" },
  },
  {
    code: "agent_machine",
    module: "agent",
    label: { zh: "机器", en: "Machines" },
    description: {
      zh: "边缘机器接入、详情与 Web Shell",
      en: "Edge machine registration, details, and the web shell",
    },
  },
  {
    code: "ssh_command",
    module: "ssh",
    label: { zh: "SSH 命令", en: "SSH commands" },
    description: {
      zh: "批量 SSH 命令执行",
      en: "Batch SSH command execution",
    },
  },
  {
    code: "web_shell",
    module: "ssh",
    label: { zh: "Web Shell", en: "Web shell" },
    description: { zh: "浏览器内终端", en: "Terminal in the browser" },
  },
]

/** 基础套餐包含的模块：前 7 个模块的全部能力（数据源和通知渠道按类型展开后共 22 项）。 */
export const BASE_MODULES = [
  "alert",
  "alert_rule",
  "probe",
  "datasource",
  "notification",
  "member",
  "role",
] as const

/** 基础套餐包含的能力，和后端 BaseFeatures() 一致。 */
export const BASE_CAPABILITY_CODES = CAPABILITIES.filter((capability) =>
  (BASE_MODULES as readonly string[]).includes(capability.module)
).map((capability) => capability.code)

/**
 * 菜单路径 → 所需能力。系统菜单（租户管理 / 套餐管理 / 用户 / 后台进程）不受
 * 能力限制，因此不在这里。
 */
export const MENU_CAPABILITIES: Record<string, string> = {
  "/overview": "alert_realtime",
  "/alerts": "alert_realtime",
  "/alert-history": "alert_history",
  "/alert-pages": "alert_page",
  "/alert-rules": "alert_rule",
  "/alert-dictionaries": "alert_dictionary",
  // 探测、数据源、通知渠道和机器按菜单授权：入口只看对应菜单。
  "/probe-targets": "probe",
  "/datasources": "datasource",
  "/notification-channels": "notification_channel",
  "/notification-templates": "notification_template",
  "/notification-groups": "notification_group",
  "/notification-subscriptions": "notification_subscription",
  "/notification-oncalls": "notification_oncall",
  "/message-logs": "message_log",
  "/members": "member",
  "/roles": "role",
  "/clusters": "cluster_access",
  "/agents": "agent",
  "/ssh-commands": "ssh_command",
}

/**
 * 授权菜单的中英标签。菜单是角色授权的一级单位，多数与能力同码；
 * 数据源、探测、通知渠道和 namespace 等合并菜单需要单独命名。
 */
export const MENU_LABELS: Record<string, LocalizedLabel> = {
  datasource: { zh: "数据源", en: "Datasources" },
  probe: { zh: "探测目标", en: "Probe targets" },
  notification_channel: { zh: "通知渠道", en: "Notification channels" },
  notification_template: { zh: "消息模板", en: "Message templates" },
  notification_group: { zh: "通知组", en: "Notification groups" },
  notification_subscription: {
    zh: "告警订阅",
    en: "Alert subscriptions",
  },
  notification_oncall: { zh: "OnCall 值班", en: "On-call" },
  message_log: { zh: "消息日志", en: "Message logs" },
  namespace: { zh: "租户设置", en: "Namespace settings" },
  agent: { zh: "机器", en: "Machines" },
}

/** 菜单标签：优先菜单目录，其次能力目录，最后回落后端名称或 code。 */
export function menuLabel(code: string, fallback?: string): LocalizedLabel {
  if (MENU_LABELS[code]) {
    return MENU_LABELS[code]
  }
  return capabilityDefinition(code)?.label ?? { zh: fallback || code, en: fallback || code }
}

/** 集群详情资源视图 → 所需能力；概况不需要能力（跟随集群接入）。 */
export const CLUSTER_SECTION_CAPABILITIES: Record<string, string> = {
  nodes: "k8s_node",
  workloads: "k8s_workload",
  pods: "k8s_pod",
  ingresses: "k8s_ingress",
  services: "k8s_service",
  configmaps: "k8s_configmap",
  secrets: "k8s_secret",
  pvs: "k8s_pv",
  pvcs: "k8s_pvc",
  jobs: "k8s_job",
  cronjobs: "k8s_cronjob",
}

/** 能力编码 → 数据源驱动。与后端 perm.datasourceCapabilityDrivers 保持一致。 */
export const DATASOURCE_CAPABILITY_DRIVERS: Record<string, string> = {
  datasource_prometheus: "METRICS_PROMETHEUS",
  datasource_victoria_metrics: "METRICS_VICTORIA_METRICS",
  datasource_elasticsearch: "LOGS_ELASTICSEARCH",
  datasource_loki: "LOGS_LOKI",
  datasource_vlogs: "LOGS_VLOGS",
  datasource_jaeger: "TRACE_JAEGER",
}

/** 能力编码 → 探测驱动。一期只有 HTTP。 */
export const PROBE_CAPABILITY_DRIVERS: Record<string, string> = {
  probe_http: "PROBE_HTTP",
}

/** 能力编码 → 通知渠道类型。 */
export const CHANNEL_CAPABILITY_TYPES: Record<string, string> = {
  notify_channel_smtp: "SMTP",
  notify_channel_webhook: "WEBHOOK",
  notify_channel_sms: "SMS",
}

/** 通知渠道类型 → 模板类型。模板是渠道的正文，渠道类型没开通时模板也不展示。 */
export const CHANNEL_TYPE_TEMPLATE_TYPES: Record<string, string[]> = {
  SMTP: ["EMAIL"],
  WEBHOOK: ["WEBHOOK"],
  SMS: ["SMS_ALICLOUD"],
}

/** 某个模块下的全部能力编码，用于把模块级菜单入口展开成任一能力。 */
export function capabilitiesOfModule(module: string): string[] {
  return CAPABILITIES.filter((item) => item.module === module).map(
    (item) => item.code
  )
}

/** 判断一个菜单能力键是模块码还是能力码。 */
export function isModuleCode(code: string): boolean {
  return MODULES.some((item) => item.code === code)
}

/** 从已开通能力里取出某项映射对应的值，保持映射声明顺序。 */
export function authorizedValues(
  features: string[],
  mapping: Record<string, string>
): string[] {
  const enabled = new Set(features)
  return Object.entries(mapping)
    .filter(([capability]) => enabled.has(capability))
    .map(([, value]) => value)
}

/** 已开通的数据源驱动。 */
export function authorizedDatasourceDrivers(features: string[]): string[] {
  return authorizedValues(features, DATASOURCE_CAPABILITY_DRIVERS)
}

/** 已开通的探测驱动。 */
export function authorizedProbeDrivers(features: string[]): string[] {
  return authorizedValues(features, PROBE_CAPABILITY_DRIVERS)
}

/** 已开通的通知渠道类型。 */
export function authorizedChannelTypes(features: string[]): string[] {
  return authorizedValues(features, CHANNEL_CAPABILITY_TYPES)
}

/** 已开通渠道类型可用的模板类型。 */
export function authorizedTemplateTypes(features: string[]): string[] {
  const types = new Set<string>()
  for (const channelType of authorizedChannelTypes(features)) {
    for (const templateType of CHANNEL_TYPE_TEMPLATE_TYPES[channelType] ?? []) {
      types.add(templateType)
    }
  }
  return [...types]
}

/**
 * 能力 → 菜单的映射。多数能力与菜单同码；按类型拆分的驱动/渠道合并到一张菜单，
 * 与后端 perm.menuSpecs 保持一致。
 */
const CAPABILITY_MENU_OVERRIDES: Record<string, string[]> = {
  datasource_prometheus: ["datasource"],
  datasource_victoria_metrics: ["datasource"],
  datasource_elasticsearch: ["datasource"],
  datasource_loki: ["datasource"],
  datasource_vlogs: ["datasource"],
  datasource_jaeger: ["datasource"],
  probe_http: ["probe"],
  notify_channel_smtp: ["notification_channel"],
  notify_channel_webhook: ["notification_channel"],
  notify_channel_sms: ["notification_channel"],
  notify_template: ["notification_template"],
  notify_group: ["notification_group"],
  notify_subscription: ["notification_subscription"],
  notify_oncall: ["notification_oncall"],
  agent_machine: ["agent"],
}

/** 返回能力开通的菜单编码。 */
export function menusOfCapability(capability: string): string[] {
  return CAPABILITY_MENU_OVERRIDES[capability] ?? [capability]
}

const MENU_CAPABILITIES_BY_CODE = (() => {
  const index: Record<string, string[]> = {}
  for (const capability of CAPABILITIES) {
    for (const menu of menusOfCapability(capability.code)) {
      index[menu] = [...(index[menu] ?? []), capability.code]
    }
  }
  return index
})()

/** 菜单入口是否可用：菜单被任一已开通能力覆盖，且角色持有该菜单权限。 */
export function menuEntryAuthorized(
  entry: string,
  features: string[],
  permissions: string[]
): boolean {
  // 租户设置是固定菜单，不受套餐能力限制。
  const licensed =
    entry === "namespace" ||
    (MENU_CAPABILITIES_BY_CODE[entry] ?? []).some((capability) =>
      features.includes(capability)
    )
  return licensed && permissions.includes(`menu:${entry}`)
}

const BY_CODE = new Map(CAPABILITIES.map((item) => [item.code, item]))
const MODULE_BY_CODE = new Map(MODULES.map((item) => [item.code, item]))
const CAPABILITY_ORDER = new Map(
  CAPABILITIES.map((capability, index) => [capability.code, index])
)
const MODULE_ORDER = new Map(MODULES.map((item, index) => [item.code, index]))

const OTHER_MODULE: LocalizedLabel = { zh: "其它", en: "Other" }

export function capabilityDefinition(
  code: string
): CapabilityDefinition | undefined {
  return BY_CODE.get(code)
}

export function moduleDefinition(code: string): ModuleDefinition | undefined {
  return MODULE_BY_CODE.get(code)
}

/** 未知能力回落到 code 本身，不空白。 */
export function capabilityLabel(code: string): LocalizedLabel {
  return BY_CODE.get(code)?.label ?? { zh: code, en: code }
}

export function moduleLabel(module: string): LocalizedLabel {
  return MODULE_BY_CODE.get(module)?.label ?? OTHER_MODULE
}

/** 当前界面语言下的能力名。 */
export function capabilityLabelText(code: string, language: string): string {
  const label = capabilityLabel(code)
  return language.startsWith("zh") ? label.zh : label.en
}

/** 次要语言标签，和 capabilityLabelText 互补。 */
export function capabilitySecondaryText(
  code: string,
  language: string
): string {
  const label = capabilityLabel(code)
  return language.startsWith("zh") ? label.en : label.zh
}

/** 「实时告警 / Realtime alerts」——没有语言上下文时的双语形态。 */
export function capabilityTitle(code: string): string {
  const label = capabilityLabel(code)
  return `${label.zh} / ${label.en}`
}

export function capabilityDescription(code: string, language: string): string {
  const description = BY_CODE.get(code)?.description
  if (!description) {
    return ""
  }
  return language.startsWith("zh") ? description.zh : description.en
}

function moduleTitleOf(label: LocalizedLabel): string {
  return `${label.zh} · ${label.en}`
}

/** 「告警 · Alerts」——模块头中英并标。 */
export function moduleTitle(module: string): string {
  return moduleTitleOf(moduleLabel(module))
}

/** 后端目录加载失败时的兜底清单。 */
export function defaultCapabilityCodes(): string[] {
  return CAPABILITIES.map((item) => item.code)
}

/** 按目录顺序排序并去重；未知 code 排在最后，避免勾选顺序漂移。 */
export function sortCapabilityCodes(codes: string[]): string[] {
  return [...new Set(codes)].sort((left, right) => {
    const leftRank = CAPABILITY_ORDER.get(left) ?? Number.MAX_SAFE_INTEGER
    const rightRank = CAPABILITY_ORDER.get(right) ?? Number.MAX_SAFE_INTEGER
    return leftRank - rightRank || left.localeCompare(right)
  })
}

export type CapabilityGroupView = {
  code: string
  title: string
  codes: string[]
}

/**
 * 按模块分组（两级勾选的第一级）。未知能力单独归入「其它」，
 * 模块顺序固定，组内按能力目录顺序。
 */
export function groupCapabilityCodes(codes: string[]): CapabilityGroupView[] {
  const grouped = new Map<string, string[]>()
  const unknown: string[] = []
  for (const code of sortCapabilityCodes(codes)) {
    const definition = BY_CODE.get(code)
    if (!definition) {
      unknown.push(code)
      continue
    }
    const list = grouped.get(definition.module)
    if (list) {
      list.push(code)
    } else {
      grouped.set(definition.module, [code])
    }
  }
  const views: CapabilityGroupView[] = MODULES.filter((module) =>
    grouped.has(module.code)
  ).map((module) => ({
    code: module.code,
    title: moduleTitleOf(module.label),
    codes: grouped.get(module.code) ?? [],
  }))
  if (unknown.length > 0) {
    views.push({
      code: "other",
      title: moduleTitleOf(OTHER_MODULE),
      codes: unknown,
    })
  }
  return views
}

/** 判断两组能力是否等价（顺序无关）。 */
export function sameCapabilitySet(
  left: Iterable<string>,
  right: Iterable<string>
): boolean {
  const a = new Set(left)
  const b = new Set(right)
  if (a.size !== b.size) {
    return false
  }
  for (const code of a) {
    if (!b.has(code)) {
      return false
    }
  }
  return true
}

/** 菜单路径需要的能力；不是能力菜单时返回 undefined。 */
export function menuCapability(path: string): string | undefined {
  return MENU_CAPABILITIES[path]
}

/** 菜单对应的菜单权限码（权限目录里 resource=menu）。 */
export function menuPermission(menu: string): string {
  return `menu:${menu}`
}

/** 菜单路径需要的菜单权限码。 */
export function menuPermissionCode(path: string): string | undefined {
  const menu = menuCapability(path)
  return menu ? menuPermission(menu) : undefined
}

/** 集群详情资源视图需要的能力；概况返回 undefined。 */
export function clusterSectionCapability(section: string): string | undefined {
  return CLUSTER_SECTION_CAPABILITIES[section]
}

/** 能力在目录里的顺序，用于权限编辑器组内排序。 */
export function capabilityRank(code: string): number {
  return CAPABILITY_ORDER.get(code) ?? Number.MAX_SAFE_INTEGER
}

/** 模块在目录里的顺序。 */
export function moduleRank(module: string): number {
  return MODULE_ORDER.get(module) ?? Number.MAX_SAFE_INTEGER
}
