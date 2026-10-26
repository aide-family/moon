import { useCallback, useMemo, useSyncExternalStore } from "react"

import { useQueryScope } from "@/hooks/use-query-scope"
import {
  createQueryTemplate,
  deleteQueryTemplate,
  clearQueryHistory,
  queryStoreVersion,
  readQueryHistory,
  readQueryTemplates,
  recordQuery,
  removeQueryHistoryEntry,
  subscribeQueryStore,
  type QueryDialect,
  type QueryHistoryEntry,
  type QuerySnapshot,
  type QueryStoreScope,
  type QueryTemplateRecord,
} from "@/lib/query-store"

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
