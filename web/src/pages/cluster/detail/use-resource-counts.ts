import { useEffect, useState } from "react"

import {
  listKubernetesConfigMaps,
  listKubernetesPersistentVolumeClaims,
  listKubernetesPersistentVolumes,
  listKubernetesSecrets,
  listKubernetesServices,
} from "@/api/cluster"

import type { ClusterNavCounts } from "./sections"

interface UseClusterResourceCountsOptions {
  clusterUid: string
  /** 留空表示全部命名空间；PV 是集群级资源，忽略该参数。 */
  namespace: string
  /** 工具栏刷新按钮递增；变化时重新取数。 */
  refreshKey?: number
}

const EMPTY_COUNTS: ClusterNavCounts = {}

// 导航计数只需要 total，pageSize=1 就够；单个接口失败时不显示该数字，不影响其它项。
async function readTotal(
  task: Promise<{ total?: number }>
): Promise<number | undefined> {
  try {
    const result = await task
    return typeof result.total === "number" ? result.total : 0
  } catch {
    return undefined
  }
}

/**
 * 服务、配置、密钥、存储这些资源视图的导航计数。
 * 它们不在概况快照里（概况只为四个既有接口做汇总），单独用轻量请求补齐。
 */
export function useClusterResourceCounts({
  clusterUid,
  namespace,
  refreshKey = 0,
}: UseClusterResourceCountsOptions): ClusterNavCounts {
  const [counts, setCounts] = useState<ClusterNavCounts>(EMPTY_COUNTS)

  useEffect(() => {
    if (!clusterUid) {
      return
    }
    let active = true

    async function load() {
      const [services, configmaps, secrets, pvs, pvcs] = await Promise.all([
        readTotal(
          listKubernetesServices(clusterUid, {
            namespace: namespace || undefined,
            page: 1,
            pageSize: 1,
          })
        ),
        readTotal(
          listKubernetesConfigMaps(clusterUid, {
            namespace: namespace || undefined,
            page: 1,
            pageSize: 1,
          })
        ),
        readTotal(
          listKubernetesSecrets(clusterUid, {
            namespace: namespace || undefined,
            page: 1,
            pageSize: 1,
          })
        ),
        readTotal(
          listKubernetesPersistentVolumes(clusterUid, {
            page: 1,
            pageSize: 1,
          })
        ),
        readTotal(
          listKubernetesPersistentVolumeClaims(clusterUid, {
            namespace: namespace || undefined,
            page: 1,
            pageSize: 1,
          })
        ),
      ])
      if (!active) {
        return
      }
      const next: ClusterNavCounts = {}
      if (services !== undefined) {
        next.services = services
      }
      if (configmaps !== undefined) {
        next.configmaps = configmaps
      }
      if (secrets !== undefined) {
        next.secrets = secrets
      }
      if (pvs !== undefined) {
        next.pvs = pvs
      }
      if (pvcs !== undefined) {
        next.pvcs = pvcs
      }
      setCounts(next)
    }

    void load()
    return () => {
      active = false
    }
  }, [clusterUid, namespace, refreshKey])

  return counts
}
