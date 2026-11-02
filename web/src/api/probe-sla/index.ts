import { get } from "@/api/http"

import type { ProbeSlaParams, ProbeSlaReport } from "./types"

/**
 * GET /v1/probe-sla
 *
 * groupUids 是重复出现的查询参数（A&groupUids=B），而通用 `get` 只拼标量，
 * 这里自己构造查询串再交给共用的请求封装（保留认证与错误本地化）。
 */
export function getProbeSla(params: ProbeSlaParams): Promise<ProbeSlaReport> {
  const query = new URLSearchParams()
  for (const uid of params.groupUids ?? []) {
    if (uid) {
      query.append("groupUids", uid)
    }
  }
  query.set("from", params.from)
  query.set("to", params.to)
  return get<ProbeSlaReport>(`/probe-sla?${query.toString()}`)
}

export type {
  ProbeSlaDay,
  ProbeSlaGroup,
  ProbeSlaParams,
  ProbeSlaReport,
} from "./types"
