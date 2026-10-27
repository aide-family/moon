import { useEffect, useRef, useState } from "react"

import {
  listKubernetesIngresses,
  listKubernetesNodes,
  listKubernetesPods,
  listKubernetesWorkloads,
} from "@/api/cluster"
import type {
  KubernetesNodeItem,
  KubernetesPodItem,
  KubernetesWorkloadItem,
} from "@/api/cluster"

import { isPodReady, parseCpuMillis, parseMemoryBytes } from "../format"
import { WORKLOAD_KINDS } from "./sections"

// Pod 列表由服务端在内存里分页，单页上限 100。概况只采样有限页，
// 避免大集群为了一个分布图把所有 Pod 拉三遍以上。
const POD_SAMPLE_PAGE_SIZE = 100
export const POD_SAMPLE_MAX_PAGES = 3
export const POD_SAMPLE_LIMIT = POD_SAMPLE_PAGE_SIZE * POD_SAMPLE_MAX_PAGES

const PHASE_ORDER = ["Running", "Pending", "Succeeded", "Failed", "Unknown"]

export interface ClusterNodeSummary {
  total: number
  ready: number
  notReady: number
  unknown: number
  cordoned: number
  cpuMillis: number
  memoryBytes: number
  pods: number
  topPods: { name: string; podCount: number }[]
}

export interface ClusterWorkloadKindSummary {
  kind: string
  total: number
  healthy: number
  desired: number
  ready: number
}

export interface ClusterWorkloadSummary {
  total: number
  healthy: number
  desired: number
  ready: number
  rows: ClusterWorkloadKindSummary[]
}

export interface ClusterPodSummary {
  total: number
  sampled: number
  truncated: boolean
  byPhase: { phase: string; count: number }[]
  notReady: number
  restarts: number
  topRestarts: { namespace: string; name: string; restarts: number }[]
}

export interface ClusterIngressSummary {
  total: number
}

export interface ClusterOverviewFailure {
  nodes: boolean
  workloads: boolean
  pods: boolean
  ingresses: boolean
}

export interface ClusterOverviewSnapshot {
  nodes: ClusterNodeSummary | null
  workloads: ClusterWorkloadSummary | null
  pods: ClusterPodSummary | null
  ingresses: ClusterIngressSummary | null
  failed: ClusterOverviewFailure
  /** 本次汇总时间（毫秒），用于展示「更新于」。 */
  refreshedAt: number
}

interface ClusterOverviewState extends ClusterOverviewSnapshot {
  signature: string
}

export interface UseClusterOverviewOptions {
  clusterUid: string
  /** 留空表示全部命名空间；节点始终是全集群口径。 */
  namespace: string
  /** 由工具栏刷新按钮递增；变化时静默重载，保留旧数据不闪空。 */
  refreshKey?: number
}

type Settled<T> = { ok: true; value: T } | { ok: false }

async function settle<T>(task: Promise<T>): Promise<Settled<T>> {
  try {
    return { ok: true, value: await task }
  } catch {
    return { ok: false }
  }
}

function readCount(value: number | undefined): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0
}

function summarizeNodes(items: KubernetesNodeItem[]): ClusterNodeSummary {
  let ready = 0
  let notReady = 0
  let cordoned = 0
  let cpuMillis = 0
  let memoryBytes = 0
  let pods = 0
  for (const node of items) {
    if (node.status === "Ready") {
      ready += 1
    } else if (node.status === "NotReady") {
      notReady += 1
    }
    if (!node.schedulable) {
      cordoned += 1
    }
    cpuMillis += parseCpuMillis(node.cpuCapacity)
    memoryBytes += parseMemoryBytes(node.memoryCapacity)
    pods += readCount(node.podCount)
  }
  const topPods = [...items]
    .sort((left, right) => readCount(right.podCount) - readCount(left.podCount))
    .slice(0, 5)
    .map((node) => ({ name: node.name, podCount: readCount(node.podCount) }))

  return {
    total: items.length,
    ready,
    notReady,
    unknown: items.length - ready - notReady,
    cordoned,
    cpuMillis,
    memoryBytes,
    pods,
    topPods,
  }
}

function workloadHealthy(item: KubernetesWorkloadItem): boolean {
  // 缩容到 0 的负载没有就绪副本也是健康的。
  return item.desiredReplicas <= 0 || item.readyReplicas >= item.desiredReplicas
}

function summarizeWorkloads(
  items: KubernetesWorkloadItem[]
): ClusterWorkloadSummary {
  const rows = WORKLOAD_KINDS.map((kind) => {
    const group = items.filter((item) => item.kind === kind)
    return {
      kind,
      total: group.length,
      healthy: group.filter(workloadHealthy).length,
      desired: group.reduce(
        (sum, item) => sum + readCount(item.desiredReplicas),
        0
      ),
      ready: group.reduce(
        (sum, item) => sum + readCount(item.readyReplicas),
        0
      ),
    }
  })
  return {
    total: items.length,
    healthy: items.filter(workloadHealthy).length,
    desired: rows.reduce((sum, row) => sum + row.desired, 0),
    ready: rows.reduce((sum, row) => sum + row.ready, 0),
    rows,
  }
}

function phaseRank(phase: string): number {
  const index = PHASE_ORDER.indexOf(phase)
  return index === -1 ? PHASE_ORDER.length : index
}

function summarizePods(
  items: KubernetesPodItem[],
  total: number
): ClusterPodSummary {
  const counts = new Map<string, number>()
  let restarts = 0
  let notReady = 0
  for (const pod of items) {
    const phase = pod.phase || "Unknown"
    counts.set(phase, (counts.get(phase) ?? 0) + 1)
    restarts += readCount(pod.restarts)
    // 已完成的 Pod 就绪字段天然是 0/N，不计入异常。
    if (phase !== "Succeeded" && phase !== "Failed" && !isPodReady(pod.ready)) {
      notReady += 1
    }
  }
  const byPhase = [...counts.entries()]
    .map(([phase, count]) => ({ phase, count }))
    .sort(
      (left, right) =>
        phaseRank(left.phase) - phaseRank(right.phase) ||
        left.phase.localeCompare(right.phase)
    )
  const topRestarts = [...items]
    .filter((pod) => readCount(pod.restarts) > 0)
    .sort((left, right) => readCount(right.restarts) - readCount(left.restarts))
    .slice(0, 5)
    .map((pod) => ({
      namespace: pod.namespace,
      name: pod.name,
      restarts: readCount(pod.restarts),
    }))

  return {
    total,
    sampled: items.length,
    truncated: total > items.length,
    byPhase,
    notReady,
    restarts,
    topRestarts,
  }
}

async function loadPodSummary(
  clusterUid: string,
  namespace: string
): Promise<ClusterPodSummary> {
  const first = await listKubernetesPods(clusterUid, {
    namespace: namespace || undefined,
    page: 1,
    pageSize: POD_SAMPLE_PAGE_SIZE,
  })
  const total = readCount(first.total) || (first.items?.length ?? 0)
  let items: KubernetesPodItem[] = first.items ?? []
  const pages = Math.min(
    Math.max(Math.ceil(total / POD_SAMPLE_PAGE_SIZE), 1),
    POD_SAMPLE_MAX_PAGES
  )
  if (pages > 1) {
    const rest = await Promise.all(
      Array.from({ length: pages - 1 }, (_, index) =>
        listKubernetesPods(clusterUid, {
          namespace: namespace || undefined,
          page: index + 2,
          pageSize: POD_SAMPLE_PAGE_SIZE,
        }).catch(() => null)
      )
    )
    for (const result of rest) {
      if (result?.items?.length) {
        items = items.concat(result.items)
      }
    }
  }
  return summarizePods(items, total)
}

/**
 * 集群概况数据：复用已有的节点、工作负载、Pod、Ingress 列表接口，
 * 在浏览器侧汇总。每个区块独立降级，单个接口失败不影响其它卡片。
 */
export function useClusterOverview({
  clusterUid,
  namespace,
  refreshKey = 0,
}: UseClusterOverviewOptions) {
  const [snapshot, setSnapshot] = useState<ClusterOverviewState | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const lastRefreshKey = useRef(refreshKey)
  const signature = `${clusterUid}::${namespace}`

  useEffect(() => {
    if (!clusterUid) {
      return
    }
    const silent = lastRefreshKey.current !== refreshKey
    lastRefreshKey.current = refreshKey
    let active = true

    async function load() {
      if (silent) {
        setRefreshing(true)
      } else {
        setLoading(true)
      }
      try {
        const [nodes, workloads, pods, ingresses] = await Promise.all([
          settle(listKubernetesNodes(clusterUid)),
          settle(
            listKubernetesWorkloads(clusterUid, {
              namespace: namespace || undefined,
            })
          ),
          settle(loadPodSummary(clusterUid, namespace)),
          settle(
            listKubernetesIngresses(clusterUid, {
              namespace: namespace || undefined,
              page: 1,
              pageSize: 1,
            })
          ),
        ])
        if (!active) {
          return
        }
        setSnapshot({
          signature,
          nodes: nodes.ok ? summarizeNodes(nodes.value.items ?? []) : null,
          workloads: workloads.ok
            ? summarizeWorkloads(workloads.value.items ?? [])
            : null,
          pods: pods.ok ? pods.value : null,
          ingresses: ingresses.ok
            ? { total: readCount(ingresses.value.total) }
            : null,
          failed: {
            nodes: !nodes.ok,
            workloads: !workloads.ok,
            pods: !pods.ok,
            ingresses: !ingresses.ok,
          },
          refreshedAt: Date.now(),
        })
      } finally {
        if (active) {
          setLoading(false)
          setRefreshing(false)
        }
      }
    }

    void load()
    return () => {
      active = false
    }
  }, [clusterUid, namespace, attempt, refreshKey, signature])

  const current = snapshot !== null && snapshot.signature === signature
  return {
    data: current ? snapshot : null,
    // 切换命名空间后旧快照不再匹配：在首帧就按加载态渲染，避免闪一帧空状态。
    loading: loading || !current,
    refreshing,
    reload: () => setAttempt((value) => value + 1),
  }
}
