export interface QueryMetricParams {
  datasourceUid: string
  expr: string
  time?: number
}

export interface QueryMetricRangeParams {
  datasourceUid: string
  expr: string
  start: number
  end: number
  step: number
}

export interface ProxyQueryParams {
  datasourceUid: string
  method: "GET" | "POST" | "PUT" | string
  path: string
  body?: string
}
