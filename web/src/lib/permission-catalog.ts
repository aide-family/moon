import type { PermissionItem } from "@/api/role/types"

// 权限字典。这里是唯一的中英文来源，故意不用 i18n 的 t()：
// 「模块 + 动作」要同时显示中文和英文，切到 en 时中文标注也不能消失。
// 字典必须与 app/alioth/pkg/perm/perm.go 的 perm.Catalog() 保持一致；
// 后端新增模块而这里没跟上时，页面回落显示原始 code，不会空白。
export type LocalizedLabel = { zh: string; en: string }

const RESOURCE_LABELS: Record<string, LocalizedLabel> = {
  member: { zh: "成员", en: "Members" },
  role: { zh: "角色", en: "Roles" },
  namespace: { zh: "租户", en: "Namespaces" },
  datasource: { zh: "数据源", en: "Datasources" },
  alert_rule: { zh: "告警规则", en: "Alert rules" },
  alert: { zh: "告警", en: "Alerts" },
  notification: { zh: "通知", en: "Notifications" },
  cluster: { zh: "Kubernetes 集群", en: "Kubernetes clusters" },
  agent: { zh: "机器", en: "Agents" },
  ssh: { zh: "SSH 命令", en: "SSH commands" },
}

const ACTION_LABELS: Record<string, LocalizedLabel> = {
  read: { zh: "读取", en: "Read" },
  write: { zh: "写入", en: "Write" },
  intervene: { zh: "干预", en: "Intervene" },
  exec: { zh: "执行", en: "Execute" },
  delete: { zh: "删除", en: "Delete" },
  transfer: { zh: "移交", en: "Transfer" },
}

export const UNKNOWN_RESOURCE: LocalizedLabel = { zh: "未分类", en: "Uncategorized" }
export const UNKNOWN_ACTION: LocalizedLabel = { zh: "未知动作", en: "Unknown action" }

/** 模块顺序与 perm.Catalog() 一致，避免同一份权限在不同页面分组顺序不同。 */
const RESOURCE_ORDER = [
  "member",
  "role",
  "namespace",
  "datasource",
  "alert_rule",
  "alert",
  "notification",
  "cluster",
  "agent",
  "ssh",
]

/** 组内顺序：读写在前，强动作在后。 */
const ACTION_ORDER = ["read", "write", "intervene", "exec", "delete", "transfer"]

/** 只属于负责人的权限，不能授予自定义角色（后端同样会拒绝）。 */
export const OWNER_ONLY = new Set(["namespace:delete", "namespace:transfer"])

export function resourceLabel(resource: string): LocalizedLabel {
  return RESOURCE_LABELS[resource] ?? { zh: resource, en: resource }
}

export function actionLabel(action: string): LocalizedLabel {
  return ACTION_LABELS[action] ?? { zh: action, en: action }
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
export function permissionTitleZh(item: Pick<PermissionItem, "resource" | "action">): string {
  return `${resourceLabel(item.resource).zh} / ${actionLabel(item.action).zh}`
}

/** 「Members / Read」——完整英文形态。 */
export function permissionTitleEn(item: Pick<PermissionItem, "resource" | "action">): string {
  return `${resourceLabel(item.resource).en} / ${actionLabel(item.action).en}`
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
  options: { assignableOnly?: boolean } = {},
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
      items: [...list].sort(
        (left, right) =>
          actionRank(left.action) - actionRank(right.action) ||
          left.name.localeCompare(right.name),
      ),
    }))
}

/** 只按 name 推导权限项，目录未加载完时详情仍能显示已授予的权限。 */
export function itemsFromNames(names: string[]): PermissionItem[] {
  return names.map((name) => {
    const [resource, ...rest] = name.split(":")
    return { uid: name, name, resource, action: rest.join(":") }
  })
}

/** 搜索命中中文、英文、模块、动作或完整 code；大小写不敏感。 */
export function matchesPermission(item: PermissionItem, query: string): boolean {
  const keyword = query.trim().toLowerCase()
  if (!keyword) {
    return true
  }
  const resource = resourceLabel(item.resource)
  const action = actionLabel(item.action)
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
