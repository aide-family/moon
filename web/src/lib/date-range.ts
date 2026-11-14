// 日期范围选择器的日期/时间运算。
// 自然日全部按本地日历计算：一天就是日历上的一天，不受夏令时 23/25 小时影响；
// 日期时间助手（parseDateTimeValue / combineDateTime 等）同样只处理本地时刻。

const DAY_MS = 86_400_000

/** 把本地自然日压成 UTC 天数序号，便于做整数天差。 */
export function dayNumber(date: Date): number {
  return Math.round(
    Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()) / DAY_MS
  )
}

function fromDayNumber(value: number): Date {
  const date = new Date(value * DAY_MS)
  return new Date(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate())
}

export function addDays(date: Date, days: number): Date {
  return fromDayNumber(dayNumber(date) + days)
}

export function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1)
}

export function addMonths(date: Date, months: number): Date {
  return new Date(date.getFullYear(), date.getMonth() + months, 1)
}

/** 两个自然日之间相差的天数（含负值），与时间戳精度无关。 */
export function diffDays(from: Date, to: Date): number {
  return dayNumber(to) - dayNumber(from)
}

/** 解析 YYYY-MM-DD 成本地自然日；非法日期（如 2026-02-30）返回 undefined。 */
export function parseDateValue(value: string): Date | undefined {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value.trim())
  if (!match) {
    return undefined
  }
  const date = new Date(
    Number(match[1]),
    Number(match[2]) - 1,
    Number(match[3])
  )
  if (
    Number.isNaN(date.getTime()) ||
    date.getMonth() !== Number(match[2]) - 1 ||
    date.getDate() !== Number(match[3])
  ) {
    return undefined
  }
  return date
}

/** 格式化成 YYYY-MM-DD，始终取本地自然日。 */
export function formatDateValue(date: Date): string {
  const month = `${date.getMonth() + 1}`.padStart(2, "0")
  const day = `${date.getDate()}`.padStart(2, "0")
  return `${date.getFullYear()}-${month}-${day}`
}

function pad2(part: number): string {
  return `${part}`.padStart(2, "0")
}

// 本地日期时间：YYYY-MM-DDTHH:mm，可带 :ss 与小数秒。
const LOCAL_DATE_TIME =
  /^(\d{4})-(\d{2})-(\d{2})[T ](\d{2}):(\d{2})(?::(\d{2})(?:\.\d+)?)?$/

/**
 * 解析本地日期时间字符串成 Date。
 * 接受 YYYY-MM-DDTHH:mm（datetime-local 的形态，可带 :ss 与小数秒），
 * 也接受带时区偏移的 RFC3339；不带时区的写法一律按本地时区解释。
 * 非法或被日历进位过的值（如 2026-02-30）返回 undefined。
 */
export function parseDateTimeValue(value: string): Date | undefined {
  const trimmed = value.trim()
  if (!trimmed) {
    return undefined
  }
  if (/^\d{4}-\d{2}-\d{2}$/.test(trimmed)) {
    // 纯日期按本地自然日的 00:00 解释，不能交给 new Date()（那样会按 UTC 解析）。
    return parseDateValue(trimmed)
  }
  const match = LOCAL_DATE_TIME.exec(trimmed)
  if (!match) {
    // 带时区偏移的 RFC3339 交给 Date 自己解析。
    const fallback = new Date(trimmed)
    return Number.isNaN(fallback.getTime()) ? undefined : fallback
  }
  const year = Number(match[1])
  const month = Number(match[2])
  const day = Number(match[3])
  const hour = Number(match[4])
  const minute = Number(match[5])
  const second = Number(match[6] ?? 0)
  const date = new Date(year, month - 1, day, hour, minute, second)
  if (
    Number.isNaN(date.getTime()) ||
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day ||
    date.getHours() !== hour ||
    date.getMinutes() !== minute ||
    date.getSeconds() !== second
  ) {
    return undefined
  }
  return date
}

/** 格式化成 YYYY-MM-DDTHH:mm（本地时区），与 datetime-local 的取值格式一致。 */
export function formatDateTimeValue(date: Date): string {
  return `${formatDateValue(date)}T${pad2(date.getHours())}:${pad2(
    date.getMinutes()
  )}`
}

/** 取日期时间值的 HH:mm；没有可解析的时间时返回空串。 */
export function timePartOf(value: string): string {
  const date = parseDateTimeValue(value)
  return date ? `${pad2(date.getHours())}:${pad2(date.getMinutes())}` : ""
}

/**
 * 把本地时刻的 HH:mm 规整成两位；非法输入返回 fallback。
 * 用于「自然日 + 时间」拼装，避免空值或半截输入混进提交值。
 */
export function normalizeTimeValue(time: string, fallback = ""): string {
  const match = /^(\d{1,2}):(\d{2})$/.exec(time.trim())
  if (!match) {
    return fallback
  }
  const hour = Number(match[1])
  const minute = Number(match[2])
  if (hour > 23 || minute > 59) {
    return fallback
  }
  return `${pad2(hour)}:${pad2(minute)}`
}

/**
 * 自然日（Date 或 YYYY-MM-DD）+ HH:mm 组合成本地 YYYY-MM-DDTHH:mm。
 * 日期或时间不合法时返回空串，调用方据此保留「未选择」语义。
 */
export function combineDateTime(day: Date | string, time: string): string {
  const date = typeof day === "string" ? parseDateValue(day) : day
  const normalized = normalizeTimeValue(time)
  if (!date || !normalized) {
    return ""
  }
  return `${formatDateValue(date)}T${normalized}`
}

/**
 * 一个月的日历格子：固定 6 行 × 7 列，周一起始。
 * 首尾会带上相邻月份的日期，保证每行都占满。
 */
export function monthDays(month: Date): Date[] {
  const first = startOfMonth(month)
  // getDay() 的周日是 0，换算成周一为 0 的偏移。
  const offset = (first.getDay() + 6) % 7
  const start = addDays(first, -offset)
  return Array.from({ length: 42 }, (_, index) => addDays(start, index))
}

// 在指定日期上补一个本地时刻，再带上本地偏移，得到不带时区歧义的 RFC3339。
function localISO(date: Date, time: string): string {
  const minutes = -date.getTimezoneOffset()
  const sign = minutes >= 0 ? "+" : "-"
  const abs = Math.abs(minutes)
  const offset = `${sign}${`${Math.floor(abs / 60)}`.padStart(2, "0")}:${`${abs % 60}`.padStart(2, "0")}`
  return `${formatDateValue(date)}T${time}${offset}`
}

/** 当日 00:00:00 的 RFC3339；作为下界使用（包含）。 */
export function dayStartISO(value: string): string | undefined {
  const date = parseDateValue(value)
  return date ? localISO(date, "00:00:00") : undefined
}

/**
 * 当日最后一刻 23:59:59.999999999 的 RFC3339；作为排除上界使用。
 * 取到纳秒是为了覆盖当天所有带小数秒的事件，效果等同次日 00:00 排除。
 */
export function dayEndISO(value: string): string | undefined {
  const date = parseDateValue(value)
  return date ? localISO(date, "23:59:59.999999999") : undefined
}

/** 当日 00:00:00 的 Unix 秒；给按秒传时间的接口用（下界包含）。 */
export function dayStartSeconds(value: string): number | undefined {
  const date = parseDateValue(value)
  return date ? Math.floor(date.getTime() / 1000) : undefined
}

/** 次日 00:00:00 的 Unix 秒；给按秒传时间的接口用（上界排除）。 */
export function nextDayStartSeconds(value: string): number | undefined {
  const date = parseDateValue(value)
  return date ? Math.floor(addDays(date, 1).getTime() / 1000) : undefined
}
