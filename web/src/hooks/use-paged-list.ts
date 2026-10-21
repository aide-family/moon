import { useEffect, useEffectEvent, useRef, useState } from "react"

import { failureText } from "@/lib/form-error"

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
  fetch: (page: number, pageSize: number) => Promise<PagedListData<T>>
  onLoaded?: (data: PagedListData<T>, meta: PagedLoadMeta) => void
}

// usePagedList 统一列表页的加载、错误、重试和分页状态。
export function usePagedList<T>({
  page,
  pageSize,
  filterKey = "",
  refreshKey = 0,
  fetch,
  onLoaded,
}: UsePagedListOptions<T>) {
  const [items, setItems] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const requestId = useRef(0)
  const lastRefreshKey = useRef(refreshKey)

  const load = useEffectEvent(async (nextPage: number, nextPageSize: number, silent: boolean) => {
    const id = ++requestId.current
    if (!silent) {
      setLoading(true)
      setFailed(false)
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
        setError(failureText(err, ""))
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
    // The effect is the synchronization point; the event reads the latest fetcher.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load(page, pageSize, false)
  }, [page, pageSize, filterKey, attempt])

  useEffect(() => {
    if (refreshKey === lastRefreshKey.current) {
      return
    }
    lastRefreshKey.current = refreshKey
    void load(page, pageSize, true)
  }, [refreshKey, page, pageSize])

  return {
    items,
    total,
    loading,
    failed,
    error,
    version: attempt,
    reload: () => setAttempt((value) => value + 1),
    fail: (message: string) => {
      setFailed(true)
      setError(message)
    },
  }
}
