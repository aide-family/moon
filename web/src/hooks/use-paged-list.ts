import { useEffect, useEffectEvent, useRef, useState } from "react"

import { ApiError } from "@/api/http"
import { failureMessage } from "@/lib/form-error"

export interface PagedListData<T> {
  items?: T[]
  total?: number
}

export interface PagedLoadMeta {
  // silent 表示后台刷新（实时轮询），不清空列表、不打断用户操作。
  silent: boolean
}

interface UsePagedListOptions<T> {
  page: number
  pageSize: number
  // filterKey 汇总当前查询条件；变化时重新加载。
  filterKey?: string
  // refreshKey 由页面定时递增；变化时静默重载，同时刷新列表和统计。
  refreshKey?: number
  // enabled=false 时暂停自动加载（例如未激活的 tab）；默认 true，其他页面行为不变。
  // 重新启用时只在页码/条件/重试序号确实变化后才加载，已加载的数据保留，
  // 因此来回切 tab 不会重复请求同一份数据。
  enabled?: boolean
  fetch: (page: number, pageSize: number) => Promise<PagedListData<T>>
  onLoaded?: (data: PagedListData<T>, meta: PagedLoadMeta) => void
}

// usePagedList 统一列表页的加载、错误、重试和分页状态。
export function usePagedList<T>({
  page,
  pageSize,
  filterKey = "",
  refreshKey = 0,
  enabled = true,
  fetch,
  onLoaded,
}: UsePagedListOptions<T>) {
  const [items, setItems] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  // 失败时保留状态码，页面据此区分网关/服务不可用（502/503）与普通列表错误。
  const [failedStatus, setFailedStatus] = useState(0)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const requestId = useRef(0)
  const lastRefreshKey = useRef(refreshKey)
  // 最近一次发起加载的查询签名；暂停后重新启用时用它判断是否需要再请求。
  const loadedSignature = useRef("")

  const load = useEffectEvent(async (nextPage: number, nextPageSize: number, silent: boolean) => {
    const id = ++requestId.current
    if (!silent) {
      setLoading(true)
      setFailed(false)
      setFailedStatus(0)
    }
    try {
      const result = await fetch(nextPage, nextPageSize)
      if (id !== requestId.current) {
        return
      }
      setItems(result.items ?? [])
      setTotal(result.total ?? 0)
      onLoaded?.(result, { silent })
    } catch (err) {
      if (id !== requestId.current) {
        return
      }
      // 静默刷新失败时保留上一次成功的数据，避免实时页闪空。
      if (!silent) {
        setFailed(true)
        setFailedStatus(err instanceof ApiError ? err.status : 0)
        setError(failureMessage(err, ""))
        setItems([])
        setTotal(0)
      }
    } finally {
      if (id === requestId.current) {
        setLoading(false)
      }
    }
  })

  useEffect(() => {
    if (!enabled) {
      return
    }
    const signature = JSON.stringify([page, pageSize, filterKey, attempt])
    // 暂停期间条件没变，恢复时直接复用已加载的数据，避免重复请求。
    if (loadedSignature.current === signature) {
      return
    }
    loadedSignature.current = signature
    // The effect is the synchronization point; the event reads the latest fetcher.
    void load(page, pageSize, false)
  }, [enabled, page, pageSize, filterKey, attempt])

  useEffect(() => {
    if (refreshKey === lastRefreshKey.current) {
      return
    }
    lastRefreshKey.current = refreshKey
    // 暂停期间跳过静默刷新；恢复时由上面的加载 effect 按需重新加载。
    if (!enabled) {
      return
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 静默刷新与首屏加载走同一套加载状态
    void load(page, pageSize, true)
  }, [enabled, refreshKey, page, pageSize])

  return {
    items,
    total,
    loading,
    failed,
    failedStatus,
    error,
    version: attempt,
    reload: () => setAttempt((value) => value + 1),
    fail: (message: string) => {
      setFailed(true)
      setFailedStatus(0)
      setError(message)
    },
  }
}
