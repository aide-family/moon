export interface VlogsLogRow {
  key: string
  timestamp: number
  stream: string
  message: string
}

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

export function toVlogsRows(
  entries: VlogsLogEntry[],
  direction: "backward" | "forward"
): VlogsLogRow[] {
  const rows = entries.map((entry, index) => {
    const time = entry._time ?? ""
    return {
      key: `${time || "row"}-${index}`,
      timestamp: Date.parse(time) || 0,
      stream: entry._stream ?? "",
      message: entry._msg ?? "",
    }
  })
  rows.sort((left, right) => {
    if (left.timestamp === right.timestamp) {
      return 0
    }
    return direction === "forward"
      ? left.timestamp - right.timestamp
      : right.timestamp - left.timestamp
  })
  return rows
}

export function formatVlogsTimestamp(timestamp: number): string {
  if (!timestamp) {
    return "—"
  }
  const date = new Date(timestamp)
  const pad = (value: number, size = 2) => String(value).padStart(size, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${pad(date.getMilliseconds(), 3)}`
}
