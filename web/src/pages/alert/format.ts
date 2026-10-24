/** 把 RFC3339 时间显示成本地时间；无法解析时保留原值。 */
export function formatTime(value: string) {
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
 * 把时长毫秒格式化成可读文本。
 * 历史事件可能持续数天，带天数才不会被换算成很大的小时数。
 */
export function formatSpan(milliseconds: number) {
  let seconds = Math.max(0, Math.floor(milliseconds / 1000))
  const days = Math.floor(seconds / 86400)
  seconds -= days * 86400
  const hours = Math.floor(seconds / 3600)
  seconds -= hours * 3600
  const minutes = Math.floor(seconds / 60)
  seconds -= minutes * 60
  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}

/**
 * 计算从开始时间到结束时间的持续时间。
 * 未结束的事件传当前时间；已结束的事件必须传 endsAt，否则时长会一直增长。
 */
export function formatDuration(start: string, endMs: number) {
  const from = new Date(start).getTime()
  if (Number.isNaN(from)) {
    return ""
  }
  return formatSpan(endMs - from)
}
