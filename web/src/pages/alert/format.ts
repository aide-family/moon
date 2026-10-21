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
 * 计算从开始时间到结束时间的持续时间。
 * 未结束的事件传当前时间，列表轮询和详情都复用这个格式。
 */
export function formatDuration(start: string, endMs: number) {
  const from = new Date(start).getTime()
  if (Number.isNaN(from)) {
    return ""
  }
  let seconds = Math.max(0, Math.floor((endMs - from) / 1000))
  const hours = Math.floor(seconds / 3600)
  seconds -= hours * 3600
  const minutes = Math.floor(seconds / 60)
  seconds -= minutes * 60
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}
