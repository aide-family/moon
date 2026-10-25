import {
  normalizeSnapshot,
  type QueryDialect,
  type QuerySnapshot,
} from "@/lib/query-store"

export type { QueryDialect, QuerySnapshot }

export const DEFAULT_ES_QUERY = '{\n  "match_all": {}\n}'

// 各方言首次进入时的默认语句。历史记录会覆盖它，但默认值仍然只在这里定义一次。
const DIALECT_DEFAULTS: Record<QueryDialect, Partial<QuerySnapshot>> = {
  promql: { expr: "up" },
  logql: { expr: '{job=~".+"}' },
  logsql: { expr: "*" },
  dsl: { expr: DEFAULT_ES_QUERY },
  http: { method: "GET", path: "/" },
}

export function defaultSnapshot(dialect: QueryDialect): QuerySnapshot {
  return normalizeSnapshot(DIALECT_DEFAULTS[dialect])
}

/** 历史列表里展示和去重用的语句：存在表达式就是表达式，HTTP 分支是「方法 路径」。 */
export function queryStatement(
  dialect: QueryDialect,
  snapshot: QuerySnapshot
): string {
  if (dialect === "http") {
    return `${snapshot.method} ${snapshot.path}`
  }
  return snapshot.expr.trim()
}

/** 单行预览：多行表达式（例如 ES 的 JSON）压成一行，避免撑高列表。 */
export function statementPreview(statement: string, maxLength = 120): string {
  const compact = statement.replace(/\s+/g, " ").trim()
  if (compact.length <= maxLength) {
    return compact
  }
  return `${compact.slice(0, maxLength)}…`
}

/** 自定义绝对时间不恢复：重新进入时应该是「最近一段」，而不是上次的时间点。 */
export function restoreSnapshot(snapshot: QuerySnapshot): QuerySnapshot {
  return snapshot.rangePreset === "custom"
    ? { ...snapshot, rangePreset: "1h" }
    : snapshot
}
