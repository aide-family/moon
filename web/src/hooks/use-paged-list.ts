import { useEffect, useEffectEvent, useState } from "react"

import { failureText } from "@/lib/form-error"

export interface PagedListData<T> {
  items?: T[]
  total?: number
}

interface UsePagedListOptions<T> {
  page: number
  pageSize: number
  // filterKey 汇总当前查询条件；变化时重新加载。
  filterKey?: string
  fetch: (page: number, pageSize: number) => Promise<PagedListData<T>>
  onLoaded?: (data: PagedListData<T>) => void
}

// usePagedList 统一列表页的加载、错误、重试和分页状态。
export function usePagedList<T>({ page, pageSize, filterKey = "", fetch, onLoaded }: UsePagedListOptions<T>) {
  const [items, setItems] = useState<T[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  const load = useEffectEvent(async (nextPage: number, nextPageSize: number) => {
    setLoading(true)
    setFailed(false)
    try {
      const result = await fetch(nextPage, nextPageSize)
      setItems(result.items ?? [])
      setTotal(result.total ?? 0)
      onLoaded?.(result)
    } catch (err) {
      setFailed(true)
      setError(failureText(err, ""))
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  })

  useEffect(() => {
    // The effect is the synchronization point; the event reads the latest fetcher.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void load(page, pageSize)
  }, [page, pageSize, filterKey, attempt])

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
