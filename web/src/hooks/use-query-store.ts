import { useCallback, useEffect, useMemo, useRef, useState, useSyncExternalStore } from "react"

import { useSession } from "@/auth/session"
import {
  createQueryTemplate,
  deleteQueryTemplate,
  normalizeSnapshot,
  clearQueryHistory,
  queryStoreVersion,
  readQueryDraft,
  readQueryHistory,
  readQueryTemplates,
  recordQuery,
  removeQueryHistoryEntry,
  subscribeQueryStore,
  writeQueryDraft,
  type QueryDialect,
  type QueryHistoryEntry,
  type QuerySnapshot,
  type QueryStoreScope,
  type QueryTemplateRecord,
} from "@/lib/query-store"

// 自定义绝对时间不落缓存：重新进入时看到的是「最近一段」，而不是上次的时间点。
function restoreDraft(
  snapshot: QuerySnapshot | null,
  defaults: QuerySnapshot
): QuerySnapshot {
  const base = normalizeSnapshot(snapshot ?? defaults)
  return base.rangePreset === "custom" ? { ...base, rangePreset: "1h" } : base
}

/** 查询缓存的作用域：用户 + 租户 + 数据源 + 方言，四者任一变化就是另一份缓存。 */
export function useQueryScope(
  datasourceUid: string,
  dialect: QueryDialect
): QueryStoreScope {
  const session = useSession()
  const userUid = session.user?.uid ?? ""
  const namespaceUid = session.namespaceUid
  return useMemo(
    () => ({ userUid, namespaceUid, datasourceUid, dialect }),
    [userUid, namespaceUid, datasourceUid, dialect]
  )
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

export interface QueryLibrary {
  scope: QueryStoreScope
  history: QueryHistoryEntry[]
  templates: QueryTemplateRecord[]
  record: (input: {
    statement: string
    snapshot: QuerySnapshot
    ok: boolean
    durationMs: number
    error?: string
  }) => void
  removeEntry: (id: string) => void
  clear: () => void
  saveTemplate: (input: {
    name: string
    description?: string
    snapshot: QuerySnapshot
  }) => void
  deleteTemplate: (id: string) => void
}

/**
 * 查询历史与自定义模板。两者都从本机存储派生，写入后通过外部存储订阅刷新，
 * 因此不需要在 effect 里同步 setState，也不会出现跨租户的旧数据。
 */
export function useQueryLibrary(
  datasourceUid: string,
  dialect: QueryDialect
): QueryLibrary {
  const scope = useQueryScope(datasourceUid, dialect)
  const version = useSyncExternalStore(subscribeQueryStore, queryStoreVersion)

  const history = useMemo(() => {
    // 版本号参与依赖：本机存储被写过就重新解析一次。
    void version
    return readQueryHistory(scope)
  }, [scope, version])
  const templates = useMemo(() => {
    void version
    return readQueryTemplates(scope)
  }, [scope, version])

  const record = useCallback<QueryLibrary["record"]>(
    (input) => {
      recordQuery(scope, input)
    },
    [scope]
  )

  const removeEntry = useCallback(
    (id: string) => {
      removeQueryHistoryEntry(scope, id)
    },
    [scope]
  )

  const clear = useCallback(() => {
    clearQueryHistory(scope)
  }, [scope])

  const saveTemplate = useCallback<QueryLibrary["saveTemplate"]>(
    (input) => {
      createQueryTemplate(scope, input)
    },
    [scope]
  )

  const deleteTemplate = useCallback(
    (id: string) => {
      deleteQueryTemplate(scope, id)
    },
    [scope]
  )

  return {
    scope,
    history,
    templates,
    record,
    removeEntry,
    clear,
    saveTemplate,
    deleteTemplate,
  }
}
