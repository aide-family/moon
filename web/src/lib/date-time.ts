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
