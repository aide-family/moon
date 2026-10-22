import { formatResultTimestamp } from "./result-data"

export type VlogsLogEntry = Record<string, string>

/** VictoriaLogs 的 /select/logsql/query 返回 JSON Lines，逐行解析，坏行直接跳过。 */
export function parseVlogsBody(body: string): VlogsLogEntry[] {
  const entries: VlogsLogEntry[] = []
  if (!body.trim()) {
    return entries
  }
  for (const line of body.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed.startsWith("{")) {
      continue
    }
    try {
      const parsed = JSON.parse(trimmed) as unknown
      if (
        typeof parsed !== "object" ||
        parsed === null ||
        Array.isArray(parsed)
      ) {
        continue
      }
      const entry: VlogsLogEntry = {}
      for (const [key, value] of Object.entries(parsed)) {
        entry[key] = typeof value === "string" ? value : JSON.stringify(value)
      }
      entries.push(entry)
    } catch {
      // 上游可能返回不完整的行，跳过即可。
    }
  }
  return entries
}

// 普通日志结果固定先展示时间、流和消息；stats 结果没有这些字段时，
// 其余字段（例如 host.name、value）按返回顺序补成表格列。
const VLOGS_LOG_FIELDS = ["_time", "_stream", "_msg"]

export function vlogsResultFields(entries: VlogsLogEntry[]): string[] {
  const seen = new Set<string>()
  const order: string[] = []
  for (const entry of entries) {
    for (const key of Object.keys(entry)) {
      if (!seen.has(key)) {
        seen.add(key)
        order.push(key)
      }
    }
  }
  const fields = VLOGS_LOG_FIELDS.filter((key) => seen.has(key))
  for (const key of order) {
    if (!fields.includes(key)) {
      fields.push(key)
    }
  }
  return fields
}

/**
 * stats 查询没有 _time 字段，不能按时间排序；普通日志仍按 _time 排序。
 * 统一做稳定排序，统计结果保持 VictoriaLogs 返回的分组顺序。
 */
export function sortVlogsEntries(
  entries: VlogsLogEntry[],
  direction: "backward" | "forward"
): VlogsLogEntry[] {
  return entries
    .map((entry, index) => ({
      entry,
      index,
      timestamp: Date.parse(entry._time ?? "") || 0,
    }))
    .sort((left, right) => {
      if (left.timestamp === right.timestamp) {
        return left.index - right.index
      }
      return direction === "forward"
        ? left.timestamp - right.timestamp
        : right.timestamp - left.timestamp
    })
    .map((item) => item.entry)
}

/** 数值列右对齐并使用等宽数字；字符串列保留原始换行。 */
export function isNumericVlogsColumn(
  entries: VlogsLogEntry[],
  field: string
): boolean {
  let found = false
  for (const entry of entries) {
    const value = entry[field]
    if (value === undefined || value === "") {
      continue
    }
    if (!/^[+-]?(?:\d+\.?\d*|\.\d+)$/.test(value.trim())) {
      return false
    }
    found = true
  }
  return found
}

export interface VlogsLabelPair {
  key: string
  value: string
}

/** VictoriaLogs 的 _stream 是一组标签，形如 {app="api", env="local"}。 */
export function parseVlogsStreamLabels(stream: string): VlogsLabelPair[] {
  const trimmed = stream.trim()
  if (!trimmed.startsWith("{") || !trimmed.endsWith("}")) {
    return []
  }
  const pairs: VlogsLabelPair[] = []
  const pattern = /([A-Za-z0-9_.-]+)\s*=\s*"((?:[^"\\]|\\.)*)"/g
  for (const match of trimmed.slice(1, -1).matchAll(pattern)) {
    pairs.push({
      key: match[1],
      value: match[2].replace(/\\(["\\])/g, "$1"),
    })
  }
  return pairs
}

/**
 * 折叠行展开后的标签：先展开 _stream 里的标签，再补充独立字段，
 * _stream 无法解析时保留原文，避免丢信息。
 */
export function vlogsLogLabels(entry: VlogsLogEntry): VlogsLabelPair[] {
  const pairs: VlogsLabelPair[] = []
  const seen = new Set<string>()
  for (const pair of parseVlogsStreamLabels(entry._stream ?? "")) {
    if (seen.has(pair.key)) {
      continue
    }
    seen.add(pair.key)
    pairs.push(pair)
  }
  for (const [key, value] of Object.entries(entry)) {
    if (key === "_time" || key === "_msg" || key === "_stream") {
      continue
    }
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    pairs.push({ key, value })
  }
  if (pairs.length === 0 && entry._stream) {
    pairs.push({ key: "_stream", value: entry._stream })
  }
  return pairs
}

const VLOGS_RESERVED_FIELDS = new Set(["_time", "_stream", "_msg"])

/** stats 结果的数值字段；_time/_stream/_msg 是日志元数据，不参与统计。 */
export function vlogsMetricFields(
  entries: VlogsLogEntry[],
  fields: string[]
): string[] {
  const numeric = fields.filter(
    (field) =>
      !VLOGS_RESERVED_FIELDS.has(field) && isNumericVlogsColumn(entries, field)
  )
  // 统计值按项目约定输出为 value；此时其余数值字段（例如按状态码分组的
  // status_code）是分组标签，不能当成第二根柱子。
  if (numeric.includes("value")) {
    return ["value"]
  }
  return numeric
}

/** 除数值字段外的其余字段，用于柱状图的分组标签。 */
export function vlogsLabelFields(
  fields: string[],
  metricFields: string[]
): string[] {
  return fields.filter((field) => !metricFields.includes(field))
}

export type VlogsResultView = "logs" | "value" | "table"

/**
 * 有 _msg 的结果按日志行展示；stats 结果的数值字段按单值或柱状图展示；
 * 其余形状（例如只返回字符串字段）保留通用表格。
 */
export function vlogsResultView(
  entries: VlogsLogEntry[],
  fields: string[]
): VlogsResultView {
  if (entries.some((entry) => "_msg" in entry)) {
    return "logs"
  }
  return vlogsMetricFields(entries, fields).length > 0 ? "value" : "table"
}

export interface VlogsStatValue {
  field: string
  raw: string
  value: number
}

export interface VlogsStatGroup {
  key: string
  label: string
  values: VlogsStatValue[]
}

/**
 * 把 stats 结果按行归组成柱状图分组：每行一个分组，
 * 每个数值字段一根柱子，多个数值字段时分组内并排。
 */
export function vlogsStatGroups(
  entries: VlogsLogEntry[],
  metricFields: string[],
  labelFields: string[]
): VlogsStatGroup[] {
  return entries.map((entry, index) => ({
    key: `${index}`,
    label: vlogsGroupLabel(entry, labelFields, index),
    values: metricFields.map((field) => {
      const raw = entry[field] ?? ""
      return { field, raw, value: raw === "" ? 0 : Number(raw) }
    }),
  }))
}

function vlogsGroupLabel(
  entry: VlogsLogEntry,
  labelFields: string[],
  index: number
): string {
  const parts: string[] = []
  for (const field of labelFields) {
    const value = entry[field] ?? ""
    if (!value) {
      continue
    }
    if (field === "_time") {
      const timestamp = Date.parse(value)
      parts.push(timestamp ? formatResultTimestamp(timestamp) : value)
      continue
    }
    parts.push(`${field}=${value}`)
  }
  return parts.length > 0 ? parts.join(" · ") : `#${index + 1}`
}

// 时间格式和柱状图布局由日志数据源共用；这里保留 VictoriaLogs 的旧导出名，
// 避免既有调用方跟着改名。
export {
  formatResultTimestamp as formatVlogsTimestamp,
  resultChartLayout as vlogsChartLayout,
} from "./result-data"
export type {
  ResultChartLayout as VlogsChartLayout,
  ResultChartPadding as VlogsChartPadding,
} from "./result-data"
