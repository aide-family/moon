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
export type InspectionTemplateFormat = "HTML" | "MARKDOWN"

export interface InspectionTemplate {
  uid: string
  name: string
  description: string
  format: InspectionTemplateFormat
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
  format: InspectionTemplateFormat
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

/** 一条报告分享链接；token 只在生成时返回一次。 */
export interface InspectionReportShare {
  uid: string
  reportUid: string
  status: "ACTIVE" | "EXPIRED" | "REVOKED"
  createdAt: string
  expiresAt: string
  revokedAt: string
  viewCount: number
  lastViewedAt: string
  hasPassword: boolean
}

/** 测试检查项返回的章节结构与 run.result.sections[] 相同。 */
export interface InspectionSection {
  title: string
  kind: InspectionItemKind
  status: "OK" | "WARN" | "FAILED"
  error?: string
  note?: string
  metrics?: InspectionMetric[]
  table?: InspectionTable
  series?: InspectionSeries[]
  logs?: InspectionLogSample[]
  alerts?: InspectionAlertRow[]
  raw?: string
}

export interface InspectionMetric {
  name: string
  value: string
  unit?: string
}

export interface InspectionTable {
  columns: string[]
  rows: string[][]
}

export interface InspectionPoint {
  timestamp: number
  value: number
}

export interface InspectionSeries {
  name: string
  points: InspectionPoint[]
}

export interface InspectionLogSample {
  time: string
  level: string
  message: string
}

export interface InspectionAlertRow {
  title: string
  level: string
  status: string
  startedAt: string
}

export interface TestInspectionItemResult {
  status: "OK" | "WARN" | "FAILED"
  error: string
  resultJson: string
  /** proto int64 在 JSON 中是字符串。 */
  elapsedMs: string | number
}

/** 指标检查项的 params。 */
export interface InspectionMetricParams {
  unit?: string
  empty_is_warn?: boolean
  thresholds?: Array<{ op: string; value: number; status: "OK" | "WARN" | "FAILED" }>
}

/** 日志检查项的 params。 */
export interface InspectionLogParams {
  levels?: string[]
  limit?: number
  direction?: "BACKWARD" | "FORWARD"
  warn_above?: number
}

/** 探测检查项的 params。 */
export interface InspectionProbeParams {
  group_uids: string[]
  metric?: "AVAILABILITY" | "LATENCY" | "P95" | "CERT_DAYS"
  availability_below?: number
  latency_above_ms?: number
  cert_days_below?: number
  warn_on_down?: boolean
}

/** 告警检查项的 params。 */
export interface InspectionAlertParams {
  status?: "FIRING" | "ALL"
  level_uids?: string[]
  labels?: string
  limit?: number
  warn_levels?: string[]
}
