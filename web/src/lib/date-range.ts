// 日期范围选择器的日期运算。
// 全部按本地自然日计算：一天就是日历上的一天，不受夏令时 23/25 小时影响。

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
