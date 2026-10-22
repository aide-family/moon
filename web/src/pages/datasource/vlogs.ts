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

export function formatVlogsTimestamp(timestamp: number): string {
  if (!timestamp) {
    return "—"
  }
  const date = new Date(timestamp)
  const pad = (value: number, size = 2) => String(value).padStart(size, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${pad(date.getMilliseconds(), 3)}`
}
