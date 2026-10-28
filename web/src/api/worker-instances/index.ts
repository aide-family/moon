import { get } from "@/api/http"

import type { ListWorkerInstancesResult, WorkerInstanceDetailResult } from "./types"

/** GET /v1/worker-instances */
export function listWorkerInstances(): Promise<ListWorkerInstancesResult> {
  return get<ListWorkerInstancesResult>("/worker-instances", undefined, undefined, { silent: true })
}

/** GET /v1/worker-instances/{instanceUid}?service=MEGREZ */
export function getWorkerInstance(
  service: string,
  instanceUid: string
): Promise<WorkerInstanceDetailResult> {
  return get<WorkerInstanceDetailResult>(
    `/worker-instances/${encodeURIComponent(instanceUid)}`,
    { service },
    undefined,
    { silent: true }
  )
}

export type {
  ListWorkerInstancesResult,
  WorkerInstanceDetail,
  WorkerInstanceDetailResult,
  WorkerInstanceItem,
  WorkerTaskItem,
} from "./types"
