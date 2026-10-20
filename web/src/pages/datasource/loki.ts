export interface LokiStream {
  stream: Record<string, string>
  values: [string, string][]
}

export interface LokiApiResponse {
  status?: string
  data?: {
    resultType?: string
    result?: LokiStream[]
  }
  error?: string
  errorType?: string
}

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

/** Loki 的时间戳是纳秒字符串。展示和排序只需要毫秒，避免 Number 转换丢失纳秒精度。 */
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
  if (!response || response.status !== "success") {
    return []
  }
  const rows: LokiLogRow[] = []
  ;(response.data?.result ?? []).forEach((stream, streamIndex) => {
    const labels = stream.stream ?? {}
    ;(stream.values ?? []).forEach(([timestamp, line], valueIndex) => {
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
