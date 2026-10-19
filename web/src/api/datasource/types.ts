export type DatasourceType = "METRICS" | "LOGS" | "TRACE"

export type DatasourceDriver =
  | "METRICS_PROMETHEUS"
  | "METRICS_VICTORIA_METRICS"
  | "LOGS_ELASTICSEARCH"
  | "TRACE_JAEGER"

export interface DatasourceItem {
  uid: string
  name: string
  type: DatasourceType | string
  driver: DatasourceDriver | string
  endpoint: string
  username: string
  secretSet: boolean
  status: string
  levelUid: string
  levelName: string
  labels: string
  remark: string
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
  username?: string
  password?: string
  secret?: string
  levelUid?: string
  remark?: string
  labels?: string
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
