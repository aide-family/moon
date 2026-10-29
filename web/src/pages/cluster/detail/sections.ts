import {
  Boxes,
  CalendarClock,
  Container,
  Database,
  Globe,
  HardDrive,
  KeyRound,
  Layers,
  LayoutDashboard,
  LayoutList,
  ListChecks,
  Network,
  Server,
  Settings2,
  type LucideIcon,
} from "lucide-react"
import type { TFunction } from "i18next"

import type { KubernetesWorkloadKind } from "@/api/cluster"

/**
 * 集群详情页的资源视图。URL 仍用 tab + kind 表达：
 * overview / nodes / workloads（kind 为空表示全部）/ jobs / cronjobs / pods /
 * ingresses / services / configmaps / secrets / pvs / pvcs。
 * Job 与 CronJob 是独立视图，忽略 kind 参数。
 */
export const CLUSTER_SECTIONS = [
  "overview",
  "nodes",
  "workloads",
  "jobs",
  "cronjobs",
  "pods",
  "ingresses",
  "services",
  "configmaps",
  "secrets",
  "pvs",
  "pvcs",
] as const

export type ClusterSection = (typeof CLUSTER_SECTIONS)[number]

export type ClusterSectionScope = "cluster" | "namespaced"

export const WORKLOAD_KINDS: readonly KubernetesWorkloadKind[] = [
  "Deployment",
  "StatefulSet",
  "DaemonSet",
]

export function isClusterSection(
  value: string | null
): value is ClusterSection {
  return (
    value !== null && (CLUSTER_SECTIONS as readonly string[]).includes(value)
  )
}

/** 导航项的展示名：专有名词（Deployment、Ingress…）直接用原文，其余走 i18n。 */
export type ClusterNavLabel = { text: string } | { i18nKey: string }

export interface ClusterNavEntry {
  /** 稳定的 React key，同时用于映射统计数字。 */
  id: string
  section: ClusterSection
  /** 仅 workloads 使用：Deployment / StatefulSet / DaemonSet；空表示全部。 */
  kind?: string
  label: ClusterNavLabel
  icon: LucideIcon
  scope: ClusterSectionScope
  /**
   * 资源级读取权限。查看该资源视图需要它；不填表示跟随页面级
   * cluster:read（概况就是这种情况）。写权限由各 Tab 单独判断。
   */
  permission?: string
}

export interface ClusterNavGroup {
  id: string
  title?: ClusterNavLabel
  entries: ClusterNavEntry[]
}

const i18n = (i18nKey: string): ClusterNavLabel => ({ i18nKey })
const text = (value: string): ClusterNavLabel => ({ text: value })

export function readClusterNavLabel(
  t: TFunction,
  label: ClusterNavLabel
): string {
  return "text" in label ? label.text : t(label.i18nKey)
}

export const CLUSTER_NAV_GROUPS: readonly ClusterNavGroup[] = [
  {
    id: "cluster",
    entries: [
      {
        id: "overview",
        section: "overview",
        label: i18n("kube.navOverview"),
        icon: LayoutDashboard,
        scope: "namespaced",
      },
      {
        id: "nodes",
        section: "nodes",
        label: i18n("kube.navNodes"),
        icon: Server,
        scope: "cluster",
        permission: "k8s:node:read",
      },
    ],
  },
  {
    id: "workloads",
    title: i18n("kube.groupWorkloads"),
    entries: [
      {
        id: "workloads",
        section: "workloads",
        label: i18n("kube.navAllWorkloads"),
        icon: LayoutList,
        scope: "namespaced",
        permission: "k8s:workload:read",
      },
      {
        id: "workloads/Deployment",
        section: "workloads",
        kind: "Deployment",
        label: text("Deployment"),
        icon: Boxes,
        scope: "namespaced",
        permission: "k8s:workload:read",
      },
      {
        id: "workloads/StatefulSet",
        section: "workloads",
        kind: "StatefulSet",
        label: text("StatefulSet"),
        icon: Database,
        scope: "namespaced",
        permission: "k8s:workload:read",
      },
      {
        id: "workloads/DaemonSet",
        section: "workloads",
        kind: "DaemonSet",
        label: text("DaemonSet"),
        icon: Layers,
        scope: "namespaced",
        permission: "k8s:workload:read",
      },
      {
        id: "jobs",
        section: "jobs",
        label: text("Job"),
        icon: ListChecks,
        scope: "namespaced",
        permission: "k8s:job:read",
      },
      {
        id: "cronjobs",
        section: "cronjobs",
        label: text("CronJob"),
        icon: CalendarClock,
        scope: "namespaced",
        permission: "k8s:cronjob:read",
      },
      {
        id: "pods",
        section: "pods",
        label: i18n("kube.navPods"),
        icon: Container,
        scope: "namespaced",
        permission: "k8s:pod:read",
      },
    ],
  },
  {
    id: "routing",
    title: i18n("kube.groupRouting"),
    entries: [
      {
        id: "ingresses",
        section: "ingresses",
        label: text("Ingress"),
        icon: Globe,
        scope: "namespaced",
        permission: "k8s:ingress:read",
      },
      {
        id: "services",
        section: "services",
        label: text("Service"),
        icon: Network,
        scope: "namespaced",
        permission: "k8s:service:read",
      },
    ],
  },
  {
    id: "config",
    title: i18n("kube.groupConfig"),
    entries: [
      {
        id: "configmaps",
        section: "configmaps",
        label: text("ConfigMap"),
        icon: Settings2,
        scope: "namespaced",
        permission: "k8s:configmap:read",
      },
      {
        id: "secrets",
        section: "secrets",
        label: text("Secret"),
        icon: KeyRound,
        scope: "namespaced",
        permission: "k8s:secret:read",
      },
    ],
  },
  {
    id: "storage",
    title: i18n("kube.groupStorage"),
    entries: [
      {
        id: "pvs",
        section: "pvs",
        label: text("PV"),
        icon: HardDrive,
        scope: "cluster",
        permission: "k8s:pv:read",
      },
      {
        id: "pvcs",
        section: "pvcs",
        label: text("PVC"),
        icon: Database,
        scope: "namespaced",
        permission: "k8s:pvc:read",
      },
    ],
  },
]

const NAV_ENTRIES: readonly ClusterNavEntry[] = CLUSTER_NAV_GROUPS.flatMap(
  (group) => group.entries
)

/** 当前 section + kind 命中的导航项；未知组合回落到 section 的第一项。 */
export function findClusterNavEntry(
  section: ClusterSection,
  kind: string
): ClusterNavEntry | null {
  const exact = NAV_ENTRIES.find(
    (entry) => entry.section === section && (entry.kind ?? "") === kind
  )
  if (exact) {
    return exact
  }
  return NAV_ENTRIES.find((entry) => entry.section === section) ?? null
}

export function clusterSectionScope(
  section: ClusterSection,
  kind: string
): ClusterSectionScope {
  return findClusterNavEntry(section, kind)?.scope ?? "namespaced"
}

/** session.can 的签名；只传判定函数，导航模块不依赖 session 实现。 */
export type ClusterPermissionCheck = (code: string) => boolean

/** 单个导航项是否可见：没有资源级权限的项跟随页面级 cluster:read。 */
export function canAccessClusterNavEntry(
  entry: ClusterNavEntry,
  can: ClusterPermissionCheck
): boolean {
  return entry.permission === undefined || can(entry.permission)
}

/** 当前 section + kind 的视图是否可访问，用于 URL 直接进入时回退到概况。 */
export function isClusterSectionAllowed(
  section: ClusterSection,
  kind: string,
  can: ClusterPermissionCheck
): boolean {
  const entry = findClusterNavEntry(section, kind)
  return entry === null || canAccessClusterNavEntry(entry, can)
}

/**
 * 按权限过滤导航：无权限的项不渲染，整组都不可见时连分组标题一起隐藏。
 * 顺序与 CLUSTER_NAV_GROUPS 一致，调用方不要自己再排序。
 */
export function filterClusterNavGroups(
  can: ClusterPermissionCheck
): ClusterNavGroup[] {
  return CLUSTER_NAV_GROUPS.map((group) => ({
    ...group,
    entries: group.entries.filter((entry) =>
      canAccessClusterNavEntry(entry, can)
    ),
  })).filter((group) => group.entries.length > 0)
}

/** 概况之外的所有资源视图都支持按名称过滤。 */
export function clusterSectionSupportsName(section: ClusterSection): boolean {
  return section !== "overview"
}

/** 导航项 id → 统计数字的映射，由概况数据和资源计数填充。 */
export type ClusterNavCounts = Record<string, number | undefined>
