export interface ElasticsearchHit {
  _index?: string
  _id?: string
  _source?: Record<string, unknown>
  sort?: unknown[]
}

export interface ElasticsearchApiResponse {
  took?: number
  timed_out?: boolean
  hits?: {
    total?: number | { value?: number; relation?: string }
    hits?: ElasticsearchHit[]
  }
  error?: unknown
  status?: number
}

export interface ElasticsearchLogRow {
  key: string
  index: string
  id: string
  timestamp: number
  source: Record<string, unknown>
}

/** 解析 Elasticsearch 的 JSON 响应。上游返回非 JSON 时返回 null，由页面展示原文。 */
export function parseElasticsearchBody(body: string): ElasticsearchApiResponse | null {
  try {
    return JSON.parse(body) as ElasticsearchApiResponse
  } catch {
    return null
  }
}

/** 按点号路径读取 _source 中的字段，例如 event.time。 */
function lookupField(source: Record<string, unknown>, path: string): unknown {
  if (path in source) {
    return source[path]
  }
  let current: unknown = source
  for (const segment of path.split(".")) {
    if (typeof current !== "object" || current === null || Array.isArray(current)) {
      return undefined
    }
    current = (current as Record<string, unknown>)[segment]
  }
  return current
}

/** Elasticsearch 的时间可能是 date 字符串、毫秒或秒级时间戳。统一转成毫秒。 */
function toMillis(value: unknown): number {
  if (typeof value === "number") {
    return value > 1e12 ? value : value * 1000
  }
  if (typeof value === "string") {
    const parsed = Date.parse(value)
    return Number.isNaN(parsed) ? 0 : parsed
  }
  return 0
}

function hitTimestamp(hit: ElasticsearchHit, timeField: string): number {
  const fromSource = toMillis(lookupField(hit._source ?? {}, timeField))
  if (fromSource > 0) {
    return fromSource
  }
  return toMillis(hit.sort?.[0])
}

/** 把命中的文档摊平成日志行。direction 决定时间顺序，和查询参数保持一致。 */
export function toElasticsearchRows(
  response: ElasticsearchApiResponse | null,
  direction: "backward" | "forward" = "backward",
  timeField = "@timestamp",
): ElasticsearchLogRow[] {
  if (!response || response.error) {
    return []
  }
  const rows = (response.hits?.hits ?? []).map((hit, index) => ({
    key: `${hit._index ?? ""}-${hit._id ?? index}`,
    index: hit._index ?? "",
    id: hit._id ?? "",
    timestamp: hitTimestamp(hit, timeField || "@timestamp"),
    source: hit._source ?? {},
  }))
  rows.sort((left, right) =>
    direction === "backward" ? right.timestamp - left.timestamp : left.timestamp - right.timestamp,
  )
  return rows
}

/** JSON 页签只展示命中文档；错误响应保留 error 与 status。 */
export function extractElasticsearchResult(response: ElasticsearchApiResponse | null): unknown {
  if (!response) {
    return null
  }
  if (response.error) {
    return { error: response.error, status: response.status }
  }
  return response.hits?.hits ?? response
}
