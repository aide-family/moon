import type { PermissionItem } from "@/api/role/types"
import {
  capabilityLabel,
  capabilityRank,
  type LocalizedLabel,
} from "@/lib/feature-catalog"

// 权限字典。这里是唯一的中英文来源，故意不用 i18n 的 t()：
// 「模块 + 动作」要同时显示中文和英文，切到 en 时中文标注也不能消失。
// 字典必须与 app/alioth/pkg/perm/perm.go 的 perm.Catalog() 保持一致；
// 后端新增模块而这里没跟上时，页面回落显示原始 code，不会空白。
// menu 模块的动作就是能力 code，标签与顺序都直接复用 feature-catalog。
export type { LocalizedLabel }

const RESOURCE_LABELS: Record<string, LocalizedLabel> = {
  // 菜单权限：动作是能力 code（menu:<能力>），决定租户菜单是否可见。
  menu: { zh: "菜单", en: "Menus" },
  member: { zh: "成员", en: "Members" },
  role: { zh: "角色", en: "Roles" },
  namespace: { zh: "租户", en: "Namespaces" },
  datasource: { zh: "数据源", en: "Datasources" },
  probe: { zh: "探测目标", en: "Probe targets" },
  alert_rule: { zh: "告警规则", en: "Alert rules" },
  alert: { zh: "告警", en: "Alerts" },
  notification: { zh: "通知", en: "Notifications" },
  cluster: { zh: "Kubernetes 集群", en: "Kubernetes clusters" },
  // cluster:* 只负责集群接入与连接；集群内具体资源按 k8s:<资源>:<动作> 授权。
  "k8s:node": { zh: "Kubernetes 节点", en: "Kubernetes nodes" },
  "k8s:workload": { zh: "Kubernetes 工作负载", en: "Kubernetes workloads" },
  "k8s:pod": { zh: "Kubernetes Pod", en: "Kubernetes pods" },
  "k8s:ingress": { zh: "Kubernetes Ingress", en: "Kubernetes ingresses" },
  "k8s:service": { zh: "Kubernetes Service", en: "Kubernetes services" },
  "k8s:configmap": { zh: "Kubernetes ConfigMap", en: "Kubernetes ConfigMaps" },
  "k8s:secret": { zh: "Kubernetes Secret", en: "Kubernetes secrets" },
  "k8s:pv": { zh: "Kubernetes PV", en: "Kubernetes persistent volumes" },
  "k8s:pvc": {
    zh: "Kubernetes PVC",
    en: "Kubernetes persistent volume claims",
  },
  "k8s:job": { zh: "Kubernetes Job", en: "Kubernetes jobs" },
  "k8s:cronjob": { zh: "Kubernetes CronJob", en: "Kubernetes cronjobs" },
  agent: { zh: "机器", en: "Agents" },
  ssh: { zh: "SSH 命令", en: "SSH commands" },
}

const ACTION_LABELS: Record<string, LocalizedLabel> = {
  read: { zh: "读取", en: "Read" },
  logs: { zh: "日志", en: "Logs" },
  write: { zh: "写入", en: "Write" },
  intervene: { zh: "干预", en: "Intervene" },
  exec: { zh: "执行", en: "Execute" },
  delete: { zh: "删除", en: "Delete" },
  transfer: { zh: "移交", en: "Transfer" },
}

export const UNKNOWN_RESOURCE: LocalizedLabel = {
  zh: "未分类",
  en: "Uncategorized",
}
export const UNKNOWN_ACTION: LocalizedLabel = {
  zh: "未知动作",
  en: "Unknown action",
}

/** 模块顺序与 perm.Catalog() 一致，避免同一份权限在不同页面分组顺序不同。 */
const RESOURCE_ORDER = [
  "member",
  "role",
  "namespace",
  "datasource",
  "probe",
  "alert_rule",
  "alert",
  "notification",
  "cluster",
  // 资源级权限紧跟在集群接入权限之后，先集群（节点 / 工作负载 / Pod），
  // 再路由、配置、存储，最后是批处理任务。
  "k8s:node",
  "k8s:workload",
  "k8s:pod",
  "k8s:ingress",
  "k8s:service",
  "k8s:configmap",
  "k8s:secret",
  "k8s:pv",
  "k8s:pvc",
  "k8s:job",
  "k8s:cronjob",
  "agent",
  "ssh",
  // 菜单权限是独立分组，排在资源权限之后：它只控制菜单是否可见，
  // 不替代资源权限，两者需要同时具备。
  "menu",
]

/** 组内顺序：读在最前（日志属于读，紧随其后），再写，最后是强动作。 */
const ACTION_ORDER = [
  "read",
  "logs",
  "write",
  "intervene",
  "exec",
  "delete",
  "transfer",
]

/** 只属于负责人的权限，不能授予自定义角色（后端同样会拒绝）。 */
export const OWNER_ONLY = new Set(["namespace:delete", "namespace:transfer"])

export function resourceLabel(resource: string): LocalizedLabel {
  return RESOURCE_LABELS[resource] ?? { zh: resource, en: resource }
}

export function actionLabel(action: string): LocalizedLabel {
  return ACTION_LABELS[action] ?? { zh: action, en: action }
}

/**
 * 权限项在组内显示的动作。menu 组的动作就是能力 code，
 * 用能力目录的双语标签渲染，不再显示原始 code。
 */
export function permissionActionLabel(
  item: Pick<PermissionItem, "resource" | "action">
): LocalizedLabel {
  if (item.resource === "menu") {
    return capabilityLabel(item.action)
  }
  return actionLabel(item.action)
}

/** 「实时告警 / Realtime alerts」——组内行内文案，中英并排。 */
export function permissionActionTitle(
  item: Pick<PermissionItem, "resource" | "action">
): string {
  const label = permissionActionLabel(item)
  return `${label.zh} / ${label.en}`
}

/** 「成员 · Members」——分组头用。 */
export function moduleTitle(resource: string): string {
  const label = resourceLabel(resource)
  return `${label.zh} · ${label.en}`
}

/** 「读取 / Read」——组内已经由分组头给出模块，行内只讲动作，中英并排。 */
export function actionTitle(action: string): string {
  const label = actionLabel(action)
  return `${label.zh} / ${label.en}`
}

/** 「成员 / 读取」——没有分组上下文时（Tooltip、审计文本）的完整中文形态。 */
export function permissionTitleZh(
  item: Pick<PermissionItem, "resource" | "action">
): string {
  return `${resourceLabel(item.resource).zh} / ${permissionActionLabel(item).zh}`
}

/** 「Members / Read」——完整英文形态。 */
export function permissionTitleEn(
  item: Pick<PermissionItem, "resource" | "action">
): string {
  return `${resourceLabel(item.resource).en} / ${permissionActionLabel(item).en}`
}

export type PermissionGroup = {
  resource: string
  title: string
  items: PermissionItem[]
}

function resourceRank(resource: string): number {
  const index = RESOURCE_ORDER.indexOf(resource)
  return index === -1 ? RESOURCE_ORDER.length : index
}

function actionRank(action: string): number {
  const index = ACTION_ORDER.indexOf(action)
  return index === -1 ? ACTION_ORDER.length : index
}

/**
 * 按模块分组。assignableOnly 为真时去掉只属于负责人的两项，
 * 这是「可分配目录」（20 项）与「展示目录」（22 项）的区别。
 */
export function groupPermissions(
  items: PermissionItem[],
  options: { assignableOnly?: boolean } = {}
): PermissionGroup[] {
  const groups = new Map<string, PermissionItem[]>()
  for (const item of items) {
    if (options.assignableOnly && OWNER_ONLY.has(item.name)) {
      continue
    }
    const list = groups.get(item.resource)
    if (list) {
      list.push(item)
    } else {
      groups.set(item.resource, [item])
    }
  }
  return [...groups.entries()]
    .sort(([left], [right]) => resourceRank(left) - resourceRank(right))
    .map(([resource, list]) => ({
      resource,
      title: moduleTitle(resource),
      items: [...list].sort((left, right) => {
        // 菜单组按能力目录顺序排，避免 alert_realtime 之类被字母序打乱。
        if (resource === "menu") {
          return (
            capabilityRank(left.action) - capabilityRank(right.action) ||
            left.name.localeCompare(right.name)
          )
        }
        return (
          actionRank(left.action) - actionRank(right.action) ||
          left.name.localeCompare(right.name)
        )
      }),
    }))
}

/** 只按 name 推导权限项，目录未加载完时详情仍能显示已授予的权限。 */
export function itemsFromNames(names: string[]): PermissionItem[] {
  return names.map((name) => {
    // 资源名本身可能含冒号（如 k8s:node:read），按已知资源前缀切分；
    // 未知资源仍回落到「第一段是资源」。
    const known = RESOURCE_ORDER.find((resource) =>
      name.startsWith(`${resource}:`)
    )
    if (known) {
      return {
        uid: name,
        name,
        resource: known,
        action: name.slice(known.length + 1),
      }
    }
    const [resource, ...rest] = name.split(":")
    return { uid: name, name, resource, action: rest.join(":") }
  })
}

/** 搜索命中中文、英文、模块、动作或完整 code；大小写不敏感。 */
export function matchesPermission(
  item: PermissionItem,
  query: string
): boolean {
  const keyword = query.trim().toLowerCase()
  if (!keyword) {
    return true
  }
  const resource = resourceLabel(item.resource)
  const action = permissionActionLabel(item)
  const haystack = [
    item.name,
    item.resource,
    item.action,
    resource.zh,
    resource.en,
    action.zh,
    action.en,
    permissionTitleZh(item),
    permissionTitleEn(item),
  ]
  return haystack.some((value) => value.toLowerCase().includes(keyword))
}
