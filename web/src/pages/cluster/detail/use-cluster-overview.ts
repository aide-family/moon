import { useEffect, useRef, useState } from "react"

import { getClusterMetricsSummary, getKubernetesOverview } from "@/api/cluster"
import type { ClusterMetricsSummary, KubernetesOverview } from "@/api/cluster"
import { failureMessage } from "@/lib/form-error"

interface OverviewSnapshot {
  signature: string
  data: KubernetesOverview | null
  failed: boolean
  error: string
  /** 本次汇总时间（毫秒），用于展示「更新于」。 */
  refreshedAt: number
}

interface MetricsSnapshot {
  signature: string
  data: ClusterMetricsSummary | null
  failed: boolean
  error: string
}

export interface ClusterMetricsState {
  data: ClusterMetricsSummary | null
  loading: boolean
  failed: boolean
  error: string
  retry: () => void
}

export interface ClusterOverviewResult {
  /** 概况聚合结果；整体请求失败时为 null，区块级降级在数组里表达。 */
  data: KubernetesOverview | null
  loading: boolean
  refreshing: boolean
  failed: boolean
  error: string
  reload: () => void
  /** 指标汇总独立加载与降级，不影响概况其余卡片。 */
  metrics: ClusterMetricsState
}

export interface UseClusterOverviewOptions {
  clusterUid: string
  /** 留空表示全部命名空间；节点与容量始终是全集群口径。 */
  namespace: string
  /** 由工具栏刷新按钮递增；变化时静默重载，保留旧数据不闪空。 */
  refreshKey?: number
}

/**
 * 集群概况数据：一次 `GET /overview` 聚合六个区块，指标汇总单独请求。
 * 两个请求互不阻塞；区块级缺权限/失败由 deniedSections / failedSections 表达。
 */
export function useClusterOverview({
  clusterUid,
  namespace,
  refreshKey = 0,
}: UseClusterOverviewOptions): ClusterOverviewResult {
  const overviewSignature = `${clusterUid}::${namespace}`
  const metricsSignature = clusterUid
  const [snapshot, setSnapshot] = useState<OverviewSnapshot | null>(null)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const lastRefreshKey = useRef(refreshKey)

  const [metricsSnapshot, setMetricsSnapshot] =
    useState<MetricsSnapshot | null>(null)
  const [metricsLoading, setMetricsLoading] = useState(true)
  const [metricsAttempt, setMetricsAttempt] = useState(0)
  const lastMetricsRefreshKey = useRef(refreshKey)

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
        const data = await getKubernetesOverview(clusterUid, {
          namespace: namespace || undefined,
        })
        if (!active) {
          return
        }
        setSnapshot({
          signature: overviewSignature,
          data,
          failed: false,
          error: "",
          refreshedAt: Date.now(),
        })
      } catch (err) {
        if (!active) {
          return
        }
        setSnapshot({
          signature: overviewSignature,
          data: null,
          failed: true,
          error: failureMessage(err, ""),
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
  }, [clusterUid, namespace, attempt, refreshKey, overviewSignature])

  useEffect(() => {
    if (!clusterUid) {
      return
    }
    const silent = lastMetricsRefreshKey.current !== refreshKey
    lastMetricsRefreshKey.current = refreshKey
    let active = true

    async function load() {
      if (!silent) {
        setMetricsLoading(true)
      }
      try {
        const data = await getClusterMetricsSummary(clusterUid)
        if (!active) {
          return
        }
        setMetricsSnapshot({
          signature: metricsSignature,
          data,
          failed: false,
          error: "",
        })
      } catch (err) {
        if (!active) {
          return
        }
        setMetricsSnapshot({
          signature: metricsSignature,
          data: null,
          failed: true,
          error: failureMessage(err, ""),
        })
      } finally {
        if (active) {
          setMetricsLoading(false)
        }
      }
    }

    void load()
    return () => {
      active = false
    }
  }, [clusterUid, metricsAttempt, refreshKey, metricsSignature])

  const current = snapshot !== null && snapshot.signature === overviewSignature
  const currentMetrics =
    metricsSnapshot !== null && metricsSnapshot.signature === metricsSignature

  return {
    data: current ? (snapshot?.data ?? null) : null,
    // 切换命名空间后旧快照不再匹配：在首帧就按加载态渲染，避免闪一帧空状态。
    loading: loading || !current,
    refreshing,
    failed: current ? (snapshot?.failed ?? false) : false,
    error: current ? (snapshot?.error ?? "") : "",
    reload: () => setAttempt((value) => value + 1),
    metrics: {
      data: currentMetrics ? (metricsSnapshot?.data ?? null) : null,
      loading: metricsLoading || !currentMetrics,
      failed: currentMetrics ? (metricsSnapshot?.failed ?? false) : false,
      error: currentMetrics ? (metricsSnapshot?.error ?? "") : "",
      retry: () => setMetricsAttempt((value) => value + 1),
    },
  }
}
