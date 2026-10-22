import type { TimeSeriesApiResponse } from "./time-series"

export interface LokiStream {
  stream: Record<string, string>
  values: [string, string][]
}

export type LokiApiResponse = TimeSeriesApiResponse

export interface LokiLogRow {
  key: string
  labels: Record<string, string>
  timestamp: number
  line: string
}

/** 解析 Loki 的 JSON 响应。上游返回非 JSON 时返回 null，由页面展示原文。 */
export function parseLokiBody(body: string): LokiApiResponse | null {
  try {
    return JSON.parse(body) as LokiApiResponse
  } catch {
    return null
  }
}

/** Loki 的 stream 是一组标签，用 Prometheus 风格的标签集展示。 */
export function formatLokiLabels(labels: Record<string, string>): string {
  const entries = Object.entries(labels ?? {})
  if (entries.length === 0) {
    return "{}"
  }
  return `{${entries.map(([key, value]) => `${key}="${value}"`).join(", ")}}`
}

function isLokiStream(value: unknown): value is LokiStream {
  if (!value || typeof value !== "object") {
    return false
  }
  const stream = value as { stream?: unknown; values?: unknown }
  if (
    stream.stream === null ||
    typeof stream.stream !== "object" ||
    Array.isArray(stream.stream) ||
    !Array.isArray(stream.values)
  ) {
    return false
  }
  return stream.values.every(
    (entry) =>
      Array.isArray(entry) &&
      entry.length >= 2 &&
      typeof entry[0] === "string" &&
      typeof entry[1] === "string",
  )
}

/** Loki 的 stream 时间戳是纳秒字符串；只转换到毫秒，避免精度丢失。 */
function nanosToMillis(nanoseconds: string): number {
  if (!/^\d+$/.test(nanoseconds) || nanoseconds.length <= 6) {
    return 0
  }
  return Number(nanoseconds.slice(0, -6))
}

/** 把多个 stream 的日志行摊平成列表。direction 决定时间顺序，和查询参数保持一致。 */
export function toLokiRows(
  response: LokiApiResponse | null,
  direction: "backward" | "forward" = "backward",
): LokiLogRow[] {
  const result = response?.data?.result
  if (
    !response ||
    response.status !== "success" ||
    response.data?.resultType !== "streams" ||
    !Array.isArray(result)
  ) {
    return []
  }
  const rows: LokiLogRow[] = []
  result.forEach((item, streamIndex) => {
    if (!isLokiStream(item)) {
      return
    }
    const { stream: labels, values } = item
    values.forEach(([timestamp, line], valueIndex) => {
      rows.push({
        key: `${streamIndex}-${valueIndex}-${timestamp}`,
        labels,
        timestamp: nanosToMillis(timestamp),
        line,
      })
    })
  })
  rows.sort((left, right) =>
    direction === "backward" ? right.timestamp - left.timestamp : left.timestamp - right.timestamp,
  )
  return rows
}

/** JSON 页签只展示结果部分，错误响应保留 error 与 errorType。 */
export function extractLokiResult(response: LokiApiResponse | null): unknown {
  if (!response) {
    return null
  }
  if (response.status === "error") {
    return { error: response.error, errorType: response.errorType }
  }
  return response.data?.result ?? response
}
