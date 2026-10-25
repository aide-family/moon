// 数据源查询的本机持久化：上次语句、查询历史、自定义模板。
// 全部按「用户 + 租户 + 数据源 + 方言」隔离，只写在 localStorage，不上传服务端。
// 读取失败（JSON 损坏、被禁用）一律当作空数据，绝不阻塞查询本身。

export type QueryDialect = "promql" | "logql" | "logsql" | "dsl" | "http"

export type QueryDirection = "backward" | "forward"

/** 一次查询的全部可恢复参数。方言只用到其中一部分，其余保持默认值。 */
export interface QuerySnapshot {
  expr: string
  rangePreset: string
  step: string
  limit: string
  direction: QueryDirection
  method: string
  path: string
  body: string
}

export const EMPTY_SNAPSHOT: QuerySnapshot = {
  expr: "",
  rangePreset: "1h",
  step: "60",
  limit: "100",
  direction: "backward",
  method: "GET",
  path: "/",
  body: "",
}

export function normalizeSnapshot(
  patch?: Partial<QuerySnapshot> | null
): QuerySnapshot {
  return { ...EMPTY_SNAPSHOT, ...(patch ?? {}) }
}

export interface QueryDraftRecord {
  snapshot: QuerySnapshot
  updatedAt: number
}

export interface QueryHistoryEntry {
  id: string
  /** 列表里展示的语句：表达式，或 HTTP 分支的「方法 路径」。 */
  statement: string
  snapshot: QuerySnapshot
  at: number
  ok: boolean
  durationMs: number
  error: string
}

export interface QueryTemplateRecord {
  id: string
  name: string
  description: string
  snapshot: QuerySnapshot
  createdAt: number
}

export interface QueryStoreScope {
  userUid: string
  namespaceUid: string
  datasourceUid: string
  dialect: QueryDialect
}

export const HISTORY_LIMIT = 50
const HISTORY_MAX_AGE_MS = 90 * 24 * 60 * 60 * 1000
const FAILED_HISTORY_MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000
const TEMPLATE_LIMIT = 50

const PREFIX = "alioth:q:v1"

// 轻量的外部存储订阅：写入后通知 React 重新读取，避免在 effect 里同步 setState。
let storeVersion = 0
const storeListeners = new Set<() => void>()

export function queryStoreVersion(): number {
  return storeVersion
}

export function subscribeQueryStore(listener: () => void): () => void {
  storeListeners.add(listener)
  return () => {
    storeListeners.delete(listener)
  }
}

function emitStoreChange() {
  storeVersion += 1
  for (const listener of storeListeners) {
    listener()
  }
}

function keyOf(scope: QueryStoreScope, kind: "draft" | "history" | "templates"): string {
  if (!scope.userUid || !scope.datasourceUid) {
    return ""
  }
  if (kind === "templates") {
    return `${PREFIX}:tpl:${scope.userUid}:${scope.namespaceUid}:${scope.dialect}`
  }
  return `${PREFIX}:${kind}:${scope.userUid}:${scope.namespaceUid}:${scope.datasourceUid}:${scope.dialect}`
}

function storage(): Storage | null {
  try {
    return window.localStorage
  } catch {
    return null
  }
}

function readJson<T>(key: string): T | null {
  const store = storage()
  if (!store || !key) {
    return null
  }
  try {
    const raw = store.getItem(key)
    if (!raw) {
      return null
    }
    return JSON.parse(raw) as T
  } catch {
    // 损坏的条目直接丢弃，避免每次进入都解析失败。
    try {
      store.removeItem(key)
    } catch {
      // 忽略：存储不可写时保持静默。
    }
    return null
  }
}

function writeJson(key: string, value: unknown): boolean {
  const store = storage()
  if (!store || !key) {
    return false
  }
  try {
    store.setItem(key, JSON.stringify(value))
    emitStoreChange()
    return true
  } catch {
    return false
  }
}

function newId(): string {
  try {
    return crypto.randomUUID()
  } catch {
    return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
  }
}

function isSnapshot(value: unknown): value is QuerySnapshot {
  if (!value || typeof value !== "object") {
    return false
  }
  const candidate = value as Partial<QuerySnapshot>
  return typeof candidate.expr === "string" && typeof candidate.rangePreset === "string"
}

export function readQueryDraft(scope: QueryStoreScope): QueryDraftRecord | null {
  const record = readJson<QueryDraftRecord>(keyOf(scope, "draft"))
  if (!record || !isSnapshot(record.snapshot)) {
    return null
  }
  return {
    snapshot: normalizeSnapshot(record.snapshot),
    updatedAt: typeof record.updatedAt === "number" ? record.updatedAt : 0,
  }
}

export function writeQueryDraft(
  scope: QueryStoreScope,
  snapshot: QuerySnapshot,
  updatedAt = Date.now()
): number {
  writeJson(keyOf(scope, "draft"), {
    snapshot: normalizeSnapshot(snapshot),
    updatedAt,
  })
  return updatedAt
}

export function readQueryHistory(scope: QueryStoreScope): QueryHistoryEntry[] {
  const list = readJson<QueryHistoryEntry[]>(keyOf(scope, "history"))
  if (!Array.isArray(list)) {
    return []
  }
  const now = Date.now()
  return list
    .filter((entry) => {
      if (!entry || typeof entry.at !== "number" || !isSnapshot(entry.snapshot)) {
        return false
      }
      const age = now - entry.at
      const limit = entry.ok ? HISTORY_MAX_AGE_MS : FAILED_HISTORY_MAX_AGE_MS
      return age <= limit
    })
    .map((entry) => ({
      id: entry.id || newId(),
      statement: typeof entry.statement === "string" ? entry.statement : "",
      snapshot: normalizeSnapshot(entry.snapshot),
      at: entry.at,
      ok: entry.ok !== false,
      durationMs: typeof entry.durationMs === "number" ? entry.durationMs : 0,
      error: typeof entry.error === "string" ? entry.error : "",
    }))
    .sort((left, right) => right.at - left.at)
    .slice(0, HISTORY_LIMIT)
}

function historySignature(statement: string, snapshot: QuerySnapshot): string {
  return `${statement}\u0000${JSON.stringify(normalizeSnapshot(snapshot))}`
}

export function recordQuery(
  scope: QueryStoreScope,
  input: {
    statement: string
    snapshot: QuerySnapshot
    ok: boolean
    durationMs: number
    error?: string
  }
): QueryHistoryEntry[] {
  const entry: QueryHistoryEntry = {
    id: newId(),
    statement: input.statement,
    snapshot: normalizeSnapshot(input.snapshot),
    at: Date.now(),
    ok: input.ok,
    durationMs: Math.max(0, Math.round(input.durationMs)),
    error: input.error ?? "",
  }
  const signature = historySignature(entry.statement, entry.snapshot)
  const rest = readQueryHistory(scope).filter(
    (item) => historySignature(item.statement, item.snapshot) !== signature
  )
  const next = [entry, ...rest].slice(0, HISTORY_LIMIT)
  writeJson(keyOf(scope, "history"), next)
  return next
}

export function removeQueryHistoryEntry(
  scope: QueryStoreScope,
  id: string
): QueryHistoryEntry[] {
  const next = readQueryHistory(scope).filter((entry) => entry.id !== id)
  writeJson(keyOf(scope, "history"), next)
  return next
}

export function clearQueryHistory(scope: QueryStoreScope): QueryHistoryEntry[] {
  writeJson(keyOf(scope, "history"), [])
  return []
}

export function readQueryTemplates(scope: QueryStoreScope): QueryTemplateRecord[] {
  const list = readJson<QueryTemplateRecord[]>(keyOf(scope, "templates"))
  if (!Array.isArray(list)) {
    return []
  }
  return list
    .filter((item) => item && typeof item.name === "string" && isSnapshot(item.snapshot))
    .map((item) => ({
      id: item.id || newId(),
      name: item.name,
      description: typeof item.description === "string" ? item.description : "",
      snapshot: normalizeSnapshot(item.snapshot),
      createdAt: typeof item.createdAt === "number" ? item.createdAt : 0,
    }))
    .sort((left, right) => right.createdAt - left.createdAt)
    .slice(0, TEMPLATE_LIMIT)
}

export function createQueryTemplate(
  scope: QueryStoreScope,
  input: { name: string; description?: string; snapshot: QuerySnapshot }
): QueryTemplateRecord[] {
  const record: QueryTemplateRecord = {
    id: newId(),
    name: input.name.trim(),
    description: (input.description ?? "").trim(),
    snapshot: normalizeSnapshot(input.snapshot),
    createdAt: Date.now(),
  }
  const next = [record, ...readQueryTemplates(scope)].slice(0, TEMPLATE_LIMIT)
  writeJson(keyOf(scope, "templates"), next)
  return next
}

export function deleteQueryTemplate(
  scope: QueryStoreScope,
  id: string
): QueryTemplateRecord[] {
  const next = readQueryTemplates(scope).filter((item) => item.id !== id)
  writeJson(keyOf(scope, "templates"), next)
  return next
}
