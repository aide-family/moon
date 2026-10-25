export type AlertEventStatus = "FIRING" | "RECOVERED" | "RECOVERED_BY_MANUAL"
export type AlertEventView = "REALTIME" | "HISTORY"

export interface AlertEventItem {
  uid: string
  ruleUid: string
  ruleName: string
  ruleGroupUid: string
  datasourceUid: string
  datasourceName: string
  fingerprint: string
  status: AlertEventStatus | string
  expr: string
  labels: string
  summary: string
  description: string
  sampleValue: string
  levelUid: string
  levelName: string
  startsAt: string
  endsAt: string
  evalSlot: string
  suppressed: boolean
  // 抑制到期时间，RFC3339；没有设置时长时为空。
  suppressedUntil: string
  intervenedBy: string
  intervenedName: string
  intervenedAt: string
  createdAt: string
}

export interface ListAlertEventsParams {
  page?: number
  pageSize?: number
  view?: AlertEventView
  name?: string
  mine?: boolean
  pageUid?: string
  levelUid?: string
  startsAtFrom?: string
  startsAtTo?: string
  eventStatus?: AlertEventStatus | string
  ruleName?: string
  datasourceName?: string
  suppression?: "SUPPRESSED" | "UNSUPPRESSED" | string
}

export interface ListAlertEventsResult {
  items: AlertEventItem[]
  total: number
  page: number
  pageSize: number
}

export interface AlertEventChart {
  datasourceName: string
  expr: string
  // protojson 会把 int64 编码成字符串，读的时候统一转成数字。
  start: number | string
  end: number | string
  step: number | string
  body: string
}

export interface AlertEventStats {
  firing: number
  suppressed: number
  recovered: number
}

export interface RealtimeAlertTabStat {
  // Tab 标识，我的告警固定为 mine，其余为告警页面 uid。
  id: string
  count: number
}

// 实时页一次刷新使用的数量统计：已选 Tab 的数量和全局数量。
export interface AlertEventRealtimeStats extends AlertEventStats {
  tabs: RealtimeAlertTabStat[]
}

export interface AlertEventUids {
  uids: string[]
}

export interface AlertEventSuppressResult {
  suppressedUntil: string
}

/** GET /v1/alert-insight 的查询参数。 */
export interface AlertInsightParams {
  /** 统计天数，1..30；0 或不传表示使用服务端默认的 7 天。 */
  days?: number
  /** 用于本地自然日分桶的 IANA 时区；空值表示 UTC。 */
  timeZone?: string
}

/** 一天的分桶。buckets 已按本地自然日升序返回，缺失日期补零。 */
export interface AlertInsightBucket {
  /** 本地自然日零点，RFC3339。 */
  start: string
  /** 当天新增事件数（按 starts_at 归日）。 */
  events: number
  /** 当天闭环数（按 ends_at 归日）。与 events 不是同一批事件，两者不构成合计。 */
  closed: number
  /** 其中人工闭环的数量。 */
  closedByManual: number
  /** 闭环时长中位数（秒）；没有样本时为 0，不能当真实值展示。 */
  p50Seconds: number
  /** 闭环时长 P90（秒），最近秩口径；没有样本时为 0。 */
  p90Seconds: number
  /** 参与分位数的闭环事件数，为 0 时 P50/P90 没有意义。 */
  recoverySamples: number
}

/** 按告警级别的分布，服务端已按事件数降序。 */
export interface AlertInsightLevel {
  levelUid: string
  events: number
  /** 其中仍未恢复的数量。 */
  firing: number
}

/** GET /v1/alert-insight 的响应（protojson，camelCase，无信封）。 */
export interface AlertInsight {
  /** 服务端实际使用的时区，分桶按它切自然日。 */
  timeZone: string
  /** 统计起始时间（含），RFC3339。 */
  from: string
  /** 统计结束时间（不含），RFC3339。 */
  to: string
  buckets: AlertInsightBucket[]
  levels: AlertInsightLevel[]
  /** 区间内闭环时长中位数（秒）。 */
  p50Seconds: number
  /** 区间内闭环时长 P90（秒）。 */
  p90Seconds: number
  /** 参与分位数的闭环事件数。 */
  recoverySamples: number
  /** 区间内新增事件总数，也是级别分布占比的分母。 */
  eventsTotal: number
  /** 区间内闭环事件总数。 */
  closedTotal: number
  /** 为假表示历史分库不可用，结果只含实时表数据。 */
  complete: boolean
}
