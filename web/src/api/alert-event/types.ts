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
