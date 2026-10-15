export type AlertEventStatus = "FIRING" | "RECOVERED" | "RECOVERED_BY_MANUAL"
export type AlertEventView = "REALTIME" | "HISTORY"

export interface AlertEventItem {
  uid: string
  ruleUid: string
  datasourceUid: string
  fingerprint: string
  status: AlertEventStatus | string
  expr: string
  labels: string
  summary: string
  description: string
  sampleValue: string
  levelUid: string
  startsAt: string
  endsAt: string
  evalSlot: string
  suppressed: boolean
  intervenedBy: string
  createdAt: string
}

export interface ListAlertEventsParams {
  page?: number
  pageSize?: number
  view?: AlertEventView
  name?: string
}

export interface ListAlertEventsResult {
  items: AlertEventItem[]
  total: number
  page: number
  pageSize: number
}

export interface AlertEventStats {
  firing: number
  suppressed: number
  recovered: number
}

export interface AlertEventUids {
  uids: string[]
}
