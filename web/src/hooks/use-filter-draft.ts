import { useState } from "react"

/**
 * 列表筛选草稿：以已生效条件（通常来自 URL）为基线，编辑期间只改草稿，
 * 提交或清除后由页面写回 URL；URL 变化（后退、外部跳转）时草稿自动回到基线。
 */
export function useFilterDraft<T extends Record<string, string>>(applied: T) {
  const signature = JSON.stringify(applied)
  const [state, setState] = useState<{
    signature: string
    draft: T
  } | null>(null)
  const draft = state?.signature === signature ? state.draft : applied

  function update<K extends keyof T>(key: K, value: T[K]) {
    setState({ signature, draft: { ...draft, [key]: value } })
  }

  function updateMany(next: Partial<T>) {
    setState({ signature, draft: { ...draft, ...next } })
  }

  function reset() {
    setState(null)
  }

  return {
    draft,
    update,
    updateMany,
    reset,
    dirty: JSON.stringify(draft) !== signature,
  }
}
