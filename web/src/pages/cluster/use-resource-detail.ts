import { useEffect, useEffectEvent, useState } from "react"

import { failureMessage } from "@/lib/form-error"

interface ClusterResourceDetailSnapshot<T> {
  /** 资源标识；与当前 props 不一致时视为尚未加载。 */
  key: string
  data: T | null
  loading: boolean
  failed: boolean
  error: string
}

interface UseClusterResourceDetailOptions<T> {
  /** 详情抽屉打开时才请求；关闭后不再发请求。 */
  enabled: boolean
  clusterUid: string
  /** 集群级资源不传。 */
  namespace?: string
  name: string
  fetch: (clusterUid: string, namespace: string, name: string) => Promise<T>
}

/**
 * 详情抽屉的关联数据：打开时加载一次，关闭后停止；身份（集群/命名空间/名称）
 * 变化或手动重试时重新加载。同一身份在重新渲染时不会重复请求。
 */
export function useClusterResourceDetail<T>({
  enabled,
  clusterUid,
  namespace = "",
  name,
  fetch,
}: UseClusterResourceDetailOptions<T>) {
  const key = `${clusterUid}\u0000${namespace}\u0000${name}`
  const [snapshot, setSnapshot] =
    useState<ClusterResourceDetailSnapshot<T> | null>(null)
  const [attempt, setAttempt] = useState(0)

  const load = useEffectEvent(async (nextKey: string) => {
    setSnapshot({
      key: nextKey,
      data: null,
      loading: true,
      failed: false,
      error: "",
    })
    try {
      const data = await fetch(clusterUid, namespace, name)
      setSnapshot({
        key: nextKey,
        data,
        loading: false,
        failed: false,
        error: "",
      })
    } catch (err) {
      setSnapshot({
        key: nextKey,
        data: null,
        loading: false,
        failed: true,
        error: failureMessage(err, ""),
      })
    }
  })

  useEffect(() => {
    if (!enabled || !clusterUid || !name) {
      return
    }
    // effect 是加载的同步点；load 通过 effect event 读取最新的 fetch。
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load(key)
  }, [enabled, clusterUid, namespace, name, attempt, key])

  const current = snapshot && snapshot.key === key ? snapshot : null
  return {
    data: current?.data ?? null,
    loading: enabled && current === null,
    failed: current?.failed ?? false,
    error: current?.error ?? "",
    retry: () => setAttempt((value) => value + 1),
  }
}
