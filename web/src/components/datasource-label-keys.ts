import { proxyQuery, queryMetric } from "@/api/metric-query"

export interface DatasourceLabelQuery {
  datasourceUid: string
  driver: string
  expression: string
}

// 补全只需要标签键名，同一数据源和表达式的结果复用，避免每次输入都请求数据源。
const cache = new Map<string, Promise<string[]>>()

/**
 * 查询规则表达式可能产出的标签键，用于摘要/描述的模板补全。指标驱动优先用
 * 表达式实查结果里的 metric 键（最贴近评测真实产出的标签），查不到再退回
 * 数据源标签接口。任何失败都返回空数组，补全退化为内置键。
 */
export function fetchDatasourceLabelKeys(
  query: DatasourceLabelQuery
): Promise<string[]> {
  const key = `${query.datasourceUid}|${query.driver}|${query.expression}`
  const cached = cache.get(key)
  if (cached) {
    return cached
  }
  const pending = probe(query).catch(() => [])
  cache.set(key, pending)
  return pending
}

async function probe(query: DatasourceLabelQuery): Promise<string[]> {
  const { datasourceUid, driver, expression } = query
  if (!datasourceUid) {
    return []
  }
  if (driver.startsWith("METRICS_") && expression.trim() !== "") {
    try {
      const response = await queryMetric({ datasourceUid, expr: expression })
      const keys = metricKeysOf(response.body)
      if (keys.length > 0) {
        return keys
      }
    } catch {
      // 表达式暂时查不出结果（数据源不可达、语法错误）时退回标签接口。
    }
  }
  if (driver.startsWith("METRICS_")) {
    const response = await proxyQuery({
      datasourceUid,
      method: "GET",
      path: "/api/v1/labels",
    })
    return parseStringList(response.body)
  }
  if (driver === "LOGS_LOKI") {
    const response = await proxyQuery({
      datasourceUid,
      method: "GET",
      path: "/loki/api/v1/labels",
    })
    return parseStringList(response.body)
  }
  if (driver === "LOGS_VLOGS") {
    const response = await proxyQuery({
      datasourceUid,
      method: "GET",
      path: `/select/logsql/field_names?query=${encodeURIComponent("_time:1d *")}&limit=500`,
    })
    return parseVlogsValueList(response.body)
  }
  return []
}

function parseJSON(body: string): unknown {
  try {
    return JSON.parse(body)
  } catch {
    return null
  }
}

function parseStringList(body: string): string[] {
  const payload = parseJSON(body) as { status?: string; data?: unknown } | null
  if (!payload || !Array.isArray(payload.data)) {
    return []
  }
  return payload.data.filter(
    (item): item is string => typeof item === "string" && item !== "__name__"
  )
}

function metricKeysOf(body: string): string[] {
  const payload = parseJSON(body) as {
    data?: { result?: Array<{ metric?: Record<string, string> }> }
  } | null
  const result = payload?.data?.result
  if (!Array.isArray(result)) {
    return []
  }
  const keys = new Set<string>()
  for (const item of result) {
    for (const key of Object.keys(item.metric ?? {})) {
      if (key !== "__name__") {
        keys.add(key)
      }
    }
  }
  return [...keys].sort()
}

function parseVlogsValueList(body: string): string[] {
  const payload = parseJSON(body) as { values?: unknown } | null
  if (!payload || !Array.isArray(payload.values)) {
    return []
  }
  const values: string[] = []
  for (const item of payload.values) {
    if (typeof item === "string") {
      values.push(item)
      continue
    }
    if (item && typeof item === "object" && "value" in item) {
      const value = (item as { value?: unknown }).value
      if (typeof value === "string") {
        values.push(value)
      }
    }
  }
  return values
}
