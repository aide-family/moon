/** 把 RFC3339 时间显示成本地日期+时间；无法解析时保留原值。 */
export function formatDateTime(value: string): string {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString()
}

/**
 * 接口里的 int64 Unix 秒经 protojson 编码后是字符串（永久套餐为 "0"）。
 * 0、负数、空串和非法值统一返回 0，表示“未设置/永久”。
 */
export function unixSeconds(value?: number | string | null): number {
  const seconds = typeof value === "number" ? value : Number(value ?? 0)
  return Number.isFinite(seconds) && seconds > 0 ? seconds : 0
}

/** Unix 秒 → 本地日期；未设置/永久返回空串。 */
export function formatUnixDate(value?: number | string | null): string {
  const seconds = unixSeconds(value)
  return seconds > 0 ? new Date(seconds * 1000).toLocaleDateString() : ""
}
