import { post } from "@/api/http"

import type { ProxyQueryParams, QueryMetricParams, QueryMetricRangeParams } from "./types"

/** POST /v1/metric-query/query */
export function queryMetric(body: QueryMetricParams): Promise<{ body: string }> {
  return post<{ body: string }>("/metric-query/query", body)
}

/** POST /v1/metric-query/query-range */
export function queryMetricRange(body: QueryMetricRangeParams): Promise<{ body: string }> {
  return post<{ body: string }>("/metric-query/query-range", body)
}

/** POST /v1/metric-query/proxy */
export function proxyQuery(body: ProxyQueryParams): Promise<{ body: string }> {
  return post<{ body: string }>("/metric-query/proxy", body)
}

export type { ProxyQueryParams, QueryMetricParams, QueryMetricRangeParams } from "./types"
