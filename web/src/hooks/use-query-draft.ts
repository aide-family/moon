import { useCallback, useEffect, useRef, useState } from "react"

import { useQueryScope } from "@/hooks/use-query-scope"
import {
  normalizeSnapshot,
  readQueryDraft,
  writeQueryDraft,
  type QueryDialect,
  type QuerySnapshot,
} from "@/lib/query-store"

// 自定义绝对时间不落缓存：重新进入时看到的是「最近一段」，而不是上次的时间点。
function restoreDraft(
  snapshot: QuerySnapshot | null,
  defaults: QuerySnapshot
): QuerySnapshot {
  const base = normalizeSnapshot(snapshot ?? defaults)
  return base.rangePreset === "custom" ? { ...base, rangePreset: "1h" } : base
}

export interface QueryDraftState {
  state: QuerySnapshot
  patch: (next: Partial<QuerySnapshot>) => void
  replace: (next: QuerySnapshot) => void
  lastUsedAt: number | null
  /** 查询发出后调用：写入上次语句，并记录「上次使用」时间。 */
  markUsed: (snapshot: QuerySnapshot) => void
}

/**
 * 上次语句回填。初始值在首次渲染前同步读取，避免先闪默认语句再被覆盖；
 * 卸载时把未执行的草稿也写回，但保留原来的「上次使用」时间。
 */
export function useQueryDraft(
  datasourceUid: string,
  dialect: QueryDialect,
  defaults: QuerySnapshot
): QueryDraftState {
  const scope = useQueryScope(datasourceUid, dialect)
  const [restored] = useState(() => readQueryDraft(scope))
  const [initial] = useState<QuerySnapshot>(() =>
    restoreDraft(restored?.snapshot ?? null, defaults)
  )
  const [state, setState] = useState<QuerySnapshot>(initial)
  const [lastUsedAt, setLastUsedAt] = useState<number | null>(
    restored?.updatedAt ? restored.updatedAt : null
  )

  const stateRef = useRef(state)
  const usedAtRef = useRef(restored?.updatedAt ?? 0)

  useEffect(() => {
    stateRef.current = state
  }, [state])

  const patch = useCallback((next: Partial<QuerySnapshot>) => {
    setState((previous) => ({ ...previous, ...next }))
  }, [])

  const replace = useCallback((next: QuerySnapshot) => {
    setState(normalizeSnapshot(next))
  }, [])

  const markUsed = useCallback(
    (snapshot: QuerySnapshot) => {
      const at = writeQueryDraft(scope, snapshot)
      usedAtRef.current = at
      setLastUsedAt(at)
    },
    [scope]
  )

  useEffect(() => {
    return () => {
      const current = stateRef.current
      // 没有改动就不写，避免无意义的「上次使用」时间。
      if (JSON.stringify(current) === JSON.stringify(initial)) {
        return
      }
      writeQueryDraft(scope, current, usedAtRef.current)
    }
  }, [scope, initial])

  return { state, patch, replace, lastUsedAt, markUsed }
}
