import { formatResultTimestamp, type ResultStatGroup } from "./result-data"
import { formatTimeSeriesName, type TimeSeriesApiResponse } from "./time-series"

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
      typeof entry[1] === "string"
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
  direction: "backward" | "forward" = "backward"
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
    direction === "backward"
      ? right.timestamp - left.timestamp
      : left.timestamp - right.timestamp
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

export type LokiResultView = "logs" | "value" | "table"

export interface LokiResultStat {
  fields: string[]
  groups: ResultStatGroup[]
}

export interface LokiTableRow {
  key: string
  series: string
  time: string
  value: string
}

interface LokiMetricSeries {
  name: string
  points: { time: number; raw: string; value: number }[]
}

interface LokiMetricItem {
  metric?: Record<string, string>
  values?: [number | string, string][]
  value?: [number | string, string]
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value)
}

function uniqueSeriesName(
  base: string,
  index: number,
  used: Set<string>
): string {
  const initial = base || `series-${index + 1}`
  let name = initial
  let suffix = 2
  while (used.has(name)) {
    name = `${initial} #${suffix}`
    suffix += 1
  }
  used.add(name)
  return name
}

function metricPoint(
  entry: unknown
): { time: number; raw: string; value: number } | null {
  if (!Array.isArray(entry) || entry.length < 2) {
    return null
  }
  const time = Number(entry[0])
  const raw = String(entry[1] ?? "")
  const value = Number.parseFloat(raw)
  if (!Number.isFinite(time) || !Number.isFinite(value)) {
    return null
  }
  return { time, raw, value }
}

/**
 * Loki 的 matrix 每个 series 有多个采样点，vector 只有最新一个点。
 * 两种形状都按 series 归一，后续柱状图按时间桶分组、按 series 画柱。
 */
function collectLokiMetricSeries(
  response: LokiApiResponse | null
): LokiMetricSeries[] {
  if (!response || response.status !== "success") {
    return []
  }
  const resultType = response.data?.resultType
  if (resultType !== "matrix" && resultType !== "vector") {
    return []
  }
  const result = response.data?.result
  if (!Array.isArray(result)) {
    return []
  }
  const used = new Set<string>()
  const series: LokiMetricSeries[] = []
  result.forEach((rawItem, index) => {
    if (!isRecord(rawItem)) {
      return
    }
    const item = rawItem as LokiMetricItem
    const metric = item.metric ?? {}
    const entries =
      resultType === "matrix"
        ? (item.values ?? [])
        : item.value
          ? [item.value]
          : []
    const points = entries
      .map(metricPoint)
      .filter((point): point is NonNullable<typeof point> => point != null)
    if (points.length === 0) {
      return
    }
    points.sort((left, right) => left.time - right.time)
    series.push({
      name: uniqueSeriesName(formatTimeSeriesName(metric), index, used),
      points,
    })
  })
  return series
}

/** 数值结果转成柱状图分组：横轴是时间桶，同一时间桶内每个 series 一根柱子。 */
export function lokiResultStat(
  response: LokiApiResponse | null
): LokiResultStat | null {
  const series = collectLokiMetricSeries(response)
  if (series.length === 0) {
    return null
  }
  const pointMaps = series.map(
    (item) => new Map(item.points.map((point) => [point.time, point]))
  )
  const timestamps = Array.from(
    new Set(series.flatMap((item) => item.points.map((point) => point.time)))
  ).sort((left, right) => left - right)
  const fields = series.map((item) => item.name)
  const groups: ResultStatGroup[] = timestamps.map((time, index) => ({
    key: `${time}-${index}`,
    label: formatResultTimestamp(time * 1000),
    values: series.map((item, seriesIndex) => {
      const point = pointMaps[seriesIndex].get(time)
      return {
        field: item.name,
        raw: point?.raw ?? "",
        value: point?.value ?? Number.NaN,
      }
    }),
  }))
  return { fields, groups }
}

/** streams 展示日志行；matrix/vector 有数值时展示单值或柱状图；其余保留表格。 */
export function lokiResultView(
  response: LokiApiResponse | null
): LokiResultView {
  if (!response || response.status !== "success") {
    return "table"
  }
  if (response.data?.resultType === "streams") {
    return "logs"
  }
  return lokiResultStat(response) ? "value" : "table"
}

/**
 * 兜底表格把 matrix/vector 摊平成 series、时间、值三列；
 * 上游返回其它结果形状时保留 JSON 文本，至少不丢信息。
 */
export function lokiTableRows(
  response: LokiApiResponse | null,
  direction: "backward" | "forward" = "backward"
): LokiTableRow[] {
  if (!response || response.status !== "success") {
    return []
  }
  const resultType = response.data?.resultType
  if (resultType === "scalar") {
    const result = response.data?.result
    if (Array.isArray(result) && result.length >= 2) {
      const time = Number(result[0])
      return [
        {
          key: `scalar-${time}`,
          series: "#1",
          time: formatResultTimestamp(time * 1000),
          value: String(result[1] ?? ""),
        },
      ]
    }
  }
  const result = response.data?.result
  if (
    (resultType === "matrix" || resultType === "vector") &&
    Array.isArray(result)
  ) {
    const used = new Set<string>()
    const rows: LokiTableRow[] = []
    result.forEach((rawItem, index) => {
      if (!isRecord(rawItem)) {
        return
      }
      const item = rawItem as LokiMetricItem
      const series = uniqueSeriesName(
        formatTimeSeriesName(item.metric ?? {}),
        index,
        used
      )
      const entries =
        resultType === "matrix"
          ? (item.values ?? [])
          : item.value
            ? [item.value]
            : []
      entries.forEach((entry, pointIndex) => {
        if (!Array.isArray(entry) || entry.length < 2) {
          return
        }
        const time = Number(entry[0])
        rows.push({
          key: `${index}-${pointIndex}-${String(entry[0])}`,
          series,
          time: Number.isFinite(time) ? formatResultTimestamp(time * 1000) : "",
          value: String(entry[1] ?? ""),
        })
      })
    })
    rows.sort((left, right) => {
      const leftTime = Date.parse(left.time) || 0
      const rightTime = Date.parse(right.time) || 0
      return direction === "backward"
        ? rightTime - leftTime
        : leftTime - rightTime
    })
    return rows
  }
  if (Array.isArray(result)) {
    return result.map((item, index) => ({
      key: `result-${index}`,
      series: `#${index + 1}`,
      time: "",
      value: typeof item === "string" ? item : JSON.stringify(item),
    }))
  }
  return []
}
