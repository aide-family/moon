export type InspectionPlanStatus = "ACTIVE" | "PAUSED"
export type InspectionItemKind =
  | "METRICS_QUERY"
  | "LOGS_QUERY"
  | "PROBE_TARGET"
  | "ALERT_EVENTS"
  | "NOTE"
export type InspectionItemDisplay =
  | "METRIC"
  | "TABLE"
  | "TREND"
  | "LOGS"
  | "TEXT"
export type InspectionOutputFormat = "HTML" | "MARKDOWN"
export type InspectionNotifyPolicy = "ALWAYS" | "ON_ABNORMAL"
export type InspectionRunStatus =
  | "PENDING"
  | "RUNNING"
  | "SUCCEEDED"
  | "PARTIAL"
  | "FAILED"
  | "CANCELED"

export interface InspectionPlanItem {
  uid: string
  kind: InspectionItemKind
  title: string
  sort: number
  datasourceUid: string
  query: string
  windowSeconds: number
  stepSeconds: number
  display: InspectionItemDisplay
  params: string
  enabled: boolean
}

export interface InspectionPlan {
  uid: string
  name: string
  description: string
  status: InspectionPlanStatus
  cron: string
  timezone: string
  timeoutSeconds: number
  templateUid: string
  outputFormats: InspectionOutputFormat[]
  titleTemplate: string
  notificationGroupUids: string[]
  notifyPolicy: InspectionNotifyPolicy
  labels: string
  remark: string
  items: InspectionPlanItem[]
  nextRunAt: string
  lastRunAt: string
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface InspectionItemInput {
  kind: InspectionItemKind
  title: string
  datasourceUid: string
  query: string
  windowSeconds: number
  stepSeconds: number
  display: InspectionItemDisplay
  params: string
  enabled: boolean
}

export interface InspectionPlanInput {
  name: string
  description: string
  cron: string
  timezone: string
  timeoutSeconds: number
  templateUid: string
  outputFormats: InspectionOutputFormat[]
  titleTemplate: string
  notificationGroupUids: string[]
  notifyPolicy: InspectionNotifyPolicy
  labels: string
  remark: string
  items: InspectionItemInput[]
}

export interface ListInspectionPlansParams {
  page?: number
  pageSize?: number
  name?: string
  status?: InspectionPlanStatus | ""
  templateUid?: string
}

export interface ListInspectionPlansResult {
  items: InspectionPlan[]
  total: number
  page: number
  pageSize: number
}

/** 报告模板：一份版式、两段正文。 */
export interface InspectionTemplate {
  uid: string
  name: string
  description: string
  htmlBody: string
  markdownBody: string
  isBuiltin: boolean
  status: "ACTIVE" | "DISABLED"
  createdAt: string
  updatedAt: string
}

export interface ListInspectionTemplatesParams {
  page?: number
  pageSize?: number
  name?: string
}

export interface ListInspectionTemplatesResult {
  items: InspectionTemplate[]
  total: number
  page: number
  pageSize: number
}

export interface SaveInspectionTemplateParams {
  name: string
  description: string
  htmlBody: string
  markdownBody: string
}

export interface InspectionRun {
  uid: string
  planUid: string
  planName: string
  trigger: "SCHEDULED" | "MANUAL"
  status: InspectionRunStatus
  scheduledAt: string
  startedAt: string
  finishedAt: string
  windowStart: string
  windowEnd: string
  itemTotal: number
  itemOk: number
  itemWarn: number
  itemFailed: number
  summary: string
  error: string
  operatorName: string
  notifiedAt: string
}

export interface ListInspectionRunsParams {
  page?: number
  pageSize?: number
  planUid?: string
  status?: InspectionRunStatus | ""
  trigger?: "SCHEDULED" | "MANUAL" | ""
  from?: string
  to?: string
}

export interface ListInspectionRunsResult {
  items: InspectionRun[]
  total: number
  page: number
  pageSize: number
}

/** 一份报告快照；列表接口不含 content。 */
export interface InspectionReport {
  uid: string
  runUid: string
  format: InspectionOutputFormat
  title: string
  content: string
  size: number
  shareEnabled: boolean
  shareExpiresAt: string
  createdAt: string
}

export interface ShareInspectionReportResult {
  token: string
  expiresAt: string
  report: InspectionReport
}
