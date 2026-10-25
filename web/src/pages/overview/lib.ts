import type { TFunction } from "i18next"

/** 本地时区的 HH:mm；无效时间显示占位符，不抛错。 */
export function formatClock(value: string): string {
  const time = Date.parse(value)
  if (Number.isNaN(time)) {
    return "—"
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(time)
}

/** 本地时区的 HH:mm:ss。 */
export function formatClockSeconds(value: string): string {
  const time = Date.parse(value)
  if (Number.isNaN(time)) {
    return "—"
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(time)
}

/**
 * 从某个时刻到现在的时长。用「天/小时/分钟」而不是相对时间短语，
 * 值班场景里「已持续 2 小时 14 分钟」比「2 小时前」更直接。
 */
export function elapsed(value: string, now: number, t: TFunction): string {
  const started = Date.parse(value)
  if (Number.isNaN(started)) {
    return "—"
  }
  const seconds = Math.max(0, Math.floor((now - started) / 1000))
  if (seconds < 60) {
    return t("overview.units.seconds", { value: seconds })
  }
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) {
    return t("overview.units.minutes", { value: minutes })
  }
  const hours = Math.floor(minutes / 60)
  if (hours < 24) {
    return t("overview.units.hours", { value: hours, rest: minutes % 60 })
  }
  return t("overview.units.days", {
    value: Math.floor(hours / 24),
    rest: hours % 24,
  })
}

export function formatCount(value: number): string {
  return new Intl.NumberFormat().format(value)
}

export function formatPercent(value: number | null, digits = 1): string {
  if (value === null || !Number.isFinite(value)) {
    return "—"
  }
  return `${(value * 100).toFixed(digits)}%`
}

/** 净增这类有正负的量：正数带 +，零就是 0。 */
export function formatSigned(value: number): string {
  if (!Number.isFinite(value) || value === 0) {
    return "0"
  }
  const text = new Intl.NumberFormat().format(Math.abs(value))
  return value > 0 ? `+${text}` : `-${text}`
}

/** 趋势图横轴：本地自然日的 MM-DD。 */
export function formatDayLabel(dateKey: string): string {
  return dateKey.slice(5)
}

/** 占比安全的百分数：总和为 0 时返回 0，避免 NaN 宽度。 */
export function share(part: number, total: number): number {
  if (total <= 0) {
    return 0
  }
  return Math.min(100, Math.max(0, (part / total) * 100))
}

/** 把最大值向上取整到 1/2/5 × 10^n，坐标轴刻度才是人读得懂的数。 */
export function niceAxisMax(value: number): number {
  if (!Number.isFinite(value) || value <= 0) {
    return 1
  }
  const exponent = Math.floor(Math.log10(value))
  const base = 10 ** exponent
  const normalized = value / base
  const step =
    normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10
  return step * base
}

/**
 * 秒数 → 可读时长。用于 P50/P90 摘要：样本为 0 时必须由调用方显示「—」，
 * 这里只负责把真实样本变成「2 分钟」这样的文案。
 */
export function formatSeconds(seconds: number, t: TFunction): string {
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "—"
  }
  if (seconds < 60) {
    return t("overview.units.seconds", { value: Math.round(seconds) })
  }
  const totalMinutes = seconds / 60
  if (totalMinutes < 60) {
    return t("overview.units.minutes", { value: Math.round(totalMinutes) })
  }
  let hours = Math.floor(totalMinutes / 60)
  let rest = Math.round(totalMinutes % 60)
  if (rest === 60) {
    hours += 1
    rest = 0
  }
  return t("overview.units.hours", { value: hours, rest })
}

/**
 * 指定时区里的「今天」（YYYY-MM-DD）。分桶用的是响应里的时区，
 * 不能用浏览器时区去匹配，否则跨时区时「今日」会错一天。
 */
export function todayKeyIn(timeZone: string, now: Date = new Date()): string {
  const pad = (value: number) => `${value}`.padStart(2, "0")
  try {
    const parts = new Intl.DateTimeFormat("en-CA", {
      timeZone,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    }).formatToParts(now)
    const pick = (type: string) =>
      parts.find((part) => part.type === type)?.value ?? ""
    const year = pick("year")
    const month = pick("month")
    const day = pick("day")
    if (year && month && day) {
      return `${year}-${month}-${day}`
    }
  } catch {
    // 非法时区名时退回浏览器本地日期。
  }
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
}

/** 桶的 start（RFC3339）→ 本地自然日键（YYYY-MM-DD），与服务端分桶对齐。 */
export function bucketDayKey(start: string): string {
  return start.slice(0, 10)
}

/**
 * 一天里闭环的数量。服务端把自动闭环和人工闭环分开返回，
 * 展示口径要合起来才是当天全部闭环。
 */
export function bucketClosed(bucket: {
  closed: number
  closedByManual: number
}): number {
  return bucket.closed + bucket.closedByManual
}
