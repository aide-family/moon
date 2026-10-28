export type DatasourceType = "METRICS" | "LOGS" | "TRACE"

export type DatasourceDriver =
  | "METRICS_PROMETHEUS"
  | "METRICS_VICTORIA_METRICS"
  | "LOGS_ELASTICSEARCH"
  | "LOGS_LOKI"
  | "LOGS_VLOGS"
  | "TRACE_JAEGER"

export interface DatasourceItem {
  uid: string
  name: string
  type: DatasourceType | string
  driver: DatasourceDriver | string
  endpoint: string
  indexPattern: string
  timeField: string
  username: string
  secretSet: boolean
  status: string
  levelUid: string
  levelName: string
  labels: string
  remark: string
  /** 是否启用数据源可用性告警。 */
  availabilityAlertEnabled: boolean
  /** 可用性告警的统计窗口，单位分钟，0..1440。 */
  failureWindowMinutes: number
  /** 统计窗口内触发可用性告警的失败次数阈值，0..1000。 */
  failureThreshold: number
  /** 可用性告警绑定的通知组标识。 */
  notificationGroupUids: string[]
  /** 绑定的通知组名称，只用于展示。 */
  notificationGroupNames: string[]
}

export interface ListDatasourcesParams {
  page?: number
  pageSize?: number
  name?: string
  type?: string
}

export interface ListDatasourcesResult {
  items: DatasourceItem[]
  total: number
  page: number
  pageSize: number
}

export interface CreateDatasourceParams {
  name: string
  type: string
  driver: string
  endpoint: string
  indexPattern?: string
  timeField?: string
  username?: string
  password?: string
  secret?: string
  levelUid?: string
  remark?: string
  labels?: string
  availabilityAlertEnabled: boolean
  failureWindowMinutes: number
  failureThreshold: number
  /** 绑定列表整体替换：空数组表示解绑全部通知组。 */
  notificationGroupUids: string[]
}

export interface UpdateDatasourceParams extends CreateDatasourceParams {
  uid: string
}

export interface SelectItem {
  uid: string
  name: string
}

export interface DatasourceMetricItem {
  name: string
  help: string
  type: string
  unit: string
}

export interface ListDatasourceMetricsResult {
  names: string[]
  metrics: DatasourceMetricItem[]
}

export interface DatasourceMetadataItem {
  key: string
  value: string
}

export interface DatasourceMetadataNode {
  name: string
  address: string
  role: string
  shards: string
  diskUsed: string
  diskAvail: string
  diskPercent: string
}

export interface DatasourceMetadataResult {
  driver: DatasourceDriver | string
  items: DatasourceMetadataItem[]
  nodes: DatasourceMetadataNode[]
}
