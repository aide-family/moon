// 功能目录（套餐 / 租户功能授权）。
//
// 功能 code 与后端 app/alioth/pkg/perm 的功能定义保持一致；GET /v1/system/features
// 返回的是同一份目录（code/name/description/permissions）。页面用这里的双语标签，
// 不直接用接口里的 name/description，这样切到 en 时中文名也不会消失，和后端
// 没部署新接口时也能先用本地目录渲染勾选框。
//
// 新增 / 删除功能时同步：后端 perm.Features()、本文件、i18n 里的菜单文案。

export type LocalizedLabel = { zh: string; en: string }

export type FeatureGroupDefinition = {
  code: string
  label: LocalizedLabel
}

export type FeatureDefinition = {
  code: string
  group: string
  label: LocalizedLabel
  description: LocalizedLabel
}

/** 功能分组顺序与权限目录一致：告警 → 通知 → 租户 → Kubernetes → 边缘。 */
export const FEATURE_GROUPS: FeatureGroupDefinition[] = [
  { code: "alert", label: { zh: "告警与监控", en: "Alerts & monitoring" } },
  { code: "notify", label: { zh: "通知", en: "Notifications" } },
  { code: "tenant", label: { zh: "租户与权限", en: "Tenant & access" } },
  { code: "kubernetes", label: { zh: "Kubernetes", en: "Kubernetes" } },
  { code: "edge", label: { zh: "边缘", en: "Edge" } },
]

/** 功能目录：顺序与后端 perm.Features() 一致，基础套餐是前 7 项。 */
export const FEATURES: FeatureDefinition[] = [
  {
    code: "alert",
    group: "alert",
    label: { zh: "告警", en: "Alerts" },
    description: {
      zh: "实时告警、历史告警与告警页面",
      en: "Realtime and historical alerts, plus alert pages",
    },
  },
  {
    code: "alert_rule",
    group: "alert",
    label: { zh: "告警规则", en: "Alert rules" },
    description: {
      zh: "告警规则与告警字典",
      en: "Alert rules and alert dictionaries",
    },
  },
  {
    code: "probe",
    group: "alert",
    label: { zh: "探测目标", en: "Probe targets" },
    description: {
      zh: "HTTP / TCP / ICMP 探测目标组",
      en: "Probe target groups for HTTP, TCP, and ICMP",
    },
  },
  {
    code: "datasource",
    group: "alert",
    label: { zh: "数据源", en: "Datasources" },
    description: {
      zh: "指标与日志数据源接入",
      en: "Metrics and log data source connections",
    },
  },
  {
    code: "notification",
    group: "notify",
    label: { zh: "通知", en: "Notifications" },
    description: {
      zh: "通知渠道、模板、通知组、告警订阅与消息日志",
      en: "Channels, templates, groups, subscriptions, and message logs",
    },
  },
  {
    code: "member",
    group: "tenant",
    label: { zh: "成员", en: "Members" },
    description: {
      zh: "租户成员管理与邀请",
      en: "Tenant membership and invitations",
    },
  },
  {
    code: "role",
    group: "tenant",
    label: { zh: "角色", en: "Roles" },
    description: {
      zh: "角色与角色权限分配",
      en: "Roles and role permission assignment",
    },
  },
  {
    code: "cluster",
    group: "kubernetes",
    label: { zh: "Kubernetes", en: "Kubernetes" },
    description: {
      zh: "Kubernetes 集群接入与资源控制台",
      en: "Kubernetes cluster access and the resource console",
    },
  },
  {
    code: "agent",
    group: "edge",
    label: { zh: "机器", en: "Machines" },
    description: {
      zh: "边缘机器接入、详情与 Web Shell",
      en: "Edge machine registration, details, and the web shell",
    },
  },
  {
    code: "ssh",
    group: "edge",
    label: { zh: "SSH 与 Web Shell", en: "SSH & web shell" },
    description: {
      zh: "批量 SSH 命令与 Web Shell",
      en: "Batch SSH commands and the web shell",
    },
  },
]

/** 基础套餐包含的功能，和后端 BaseFeatures() 一致。 */
export const BASE_FEATURE_CODES = [
  "alert",
  "alert_rule",
  "probe",
  "datasource",
  "notification",
  "member",
  "role",
] as const

const BY_CODE = new Map(FEATURES.map((feature) => [feature.code, feature]))
const GROUP_BY_CODE = new Map(
  FEATURE_GROUPS.map((group) => [group.code, group])
)
const FEATURE_ORDER = new Map(
  FEATURES.map((feature, index) => [feature.code, index])
)

const OTHER_GROUP: LocalizedLabel = { zh: "其它", en: "Other" }

export function featureDefinition(code: string): FeatureDefinition | undefined {
  return BY_CODE.get(code)
}

/** 未知功能回落到 code 本身，不空白。 */
export function featureLabel(code: string): LocalizedLabel {
  return BY_CODE.get(code)?.label ?? { zh: code, en: code }
}

/** 「告警 · Alerts」——按当前界面语言优先，另一种语言作为次要标注。 */
export function featureLabelText(code: string, language: string): string {
  const label = featureLabel(code)
  return language.startsWith("zh") ? label.zh : label.en
}

/** 次要语言标签，和 featureLabelText 互补。 */
export function featureSecondaryText(code: string, language: string): string {
  const label = featureLabel(code)
  return language.startsWith("zh") ? label.en : label.zh
}

/** 「告警 / Alerts」——没有语言上下文时的双语形态（标签 title、审计文本）。 */
export function featureTitle(code: string): string {
  const label = featureLabel(code)
  return `${label.zh} / ${label.en}`
}

export function featureDescription(code: string, language: string): string {
  const description = BY_CODE.get(code)?.description
  if (!description) {
    return ""
  }
  return language.startsWith("zh") ? description.zh : description.en
}

export function featureGroupLabel(group: string): LocalizedLabel {
  return GROUP_BY_CODE.get(group)?.label ?? OTHER_GROUP
}

/** 「告警与监控 · Alerts & monitoring」——分组头中英并标。 */
export function featureGroupTitle(group: string): string {
  const label = featureGroupLabel(group)
  return `${label.zh} · ${label.en}`
}

/** 后端功能目录加载失败时的兜底清单。 */
export function defaultFeatureCodes(): string[] {
  return FEATURES.map((feature) => feature.code)
}

/** 按后端目录顺序排序并去重；未知 code 排在最后，避免勾选顺序漂移。 */
export function sortFeatureCodes(codes: string[]): string[] {
  return [...new Set(codes)].sort((left, right) => {
    const leftRank = FEATURE_ORDER.get(left) ?? Number.MAX_SAFE_INTEGER
    const rightRank = FEATURE_ORDER.get(right) ?? Number.MAX_SAFE_INTEGER
    return leftRank - rightRank || left.localeCompare(right)
  })
}

export type FeatureGroupView = {
  code: string
  title: string
  codes: string[]
}

function groupTitle(label: LocalizedLabel): string {
  return `${label.zh} · ${label.en}`
}

/** 按分组展示功能；未知功能单独归入「其它」。 */
export function groupFeatureCodes(codes: string[]): FeatureGroupView[] {
  const grouped = new Map<string, string[]>()
  const unknown: string[] = []
  for (const code of sortFeatureCodes(codes)) {
    const definition = BY_CODE.get(code)
    if (!definition) {
      unknown.push(code)
      continue
    }
    const list = grouped.get(definition.group)
    if (list) {
      list.push(code)
    } else {
      grouped.set(definition.group, [code])
    }
  }
  const views: FeatureGroupView[] = FEATURE_GROUPS.filter((group) =>
    grouped.has(group.code)
  ).map((group) => ({
    code: group.code,
    title: groupTitle(group.label),
    codes: grouped.get(group.code) ?? [],
  }))
  if (unknown.length > 0) {
    views.push({
      code: "other",
      title: groupTitle(OTHER_GROUP),
      codes: unknown,
    })
  }
  return views
}

/** 判断两组功能是否等价（顺序无关）。 */
export function sameFeatureSet(
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
