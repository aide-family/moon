import { get } from "@/api/http"

import type { ListWorkerInstancesResult } from "./types"

/** GET /v1/worker-instances */
export function listWorkerInstances(): Promise<ListWorkerInstancesResult> {
  return get<ListWorkerInstancesResult>("/worker-instances", undefined, undefined, { silent: true })
}

export type { ListWorkerInstancesResult, WorkerInstanceItem } from "./types"
