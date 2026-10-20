import { post } from "@/api/http"

import type { QueryLogsParams } from "./types"

/** POST /v1/logs-query/query */
export function queryLogs(body: QueryLogsParams): Promise<{ body: string }> {
  return post<{ body: string }>("/logs-query/query", body)
}

export type { QueryLogsParams } from "./types"
