import { del, get, post, put } from "@/api/http"

import type {
  CreateProbeTargetGroupParams,
  ListProbeTargetGroupsParams,
  ListProbeTargetGroupsResult,
  ListProbeTargetResultsParams,
  ListProbeTargetResultsResult,
  ProbeResultItem,
  ProbeTargetGroupItem,
  ProbeTrialParams,
  ProbeTrialResult,
  UpdateProbeTargetGroupParams,
} from "./types"

/** GET /v1/probe-target-groups */
export function listProbeTargetGroups(
  params: ListProbeTargetGroupsParams = {}
): Promise<ListProbeTargetGroupsResult> {
  return get<ListProbeTargetGroupsResult>("/probe-target-groups", params)
}

/** GET /v1/probe-target-groups/{uid} */
export function getProbeTargetGroup(
  uid: string
): Promise<{ group: ProbeTargetGroupItem }> {
  return get<{ group: ProbeTargetGroupItem }>(
    `/probe-target-groups/${uid}`
  )
}

/** POST /v1/probe-target-groups */
export function createProbeTargetGroup(
  body: CreateProbeTargetGroupParams
): Promise<{ group: ProbeTargetGroupItem }> {
  return post<{ group: ProbeTargetGroupItem }>(
    "/probe-target-groups",
    body
  )
}

/** PUT /v1/probe-target-groups/{uid} */
export function updateProbeTargetGroup(
  body: UpdateProbeTargetGroupParams
): Promise<{ group: ProbeTargetGroupItem }> {
  return put<{ group: ProbeTargetGroupItem }>(
    `/probe-target-groups/${body.uid}`,
    body
  )
}

/** PUT /v1/probe-target-groups/{uid}/status */
export function updateProbeTargetGroupStatus(
  uid: string,
  status: string
): Promise<{ group: ProbeTargetGroupItem }> {
  return put<{ group: ProbeTargetGroupItem }>(
    `/probe-target-groups/${uid}/status`,
    { uid, status }
  )
}

/** DELETE /v1/probe-target-groups/{uid} */
export function deleteProbeTargetGroup(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/probe-target-groups/${uid}`)
}

/** GET /v1/probe-target-groups/{uid}/results */
export function listProbeTargetResults(
  uid: string,
  params: ListProbeTargetResultsParams = {}
): Promise<ListProbeTargetResultsResult> {
  return get<ListProbeTargetResultsResult>(
    `/probe-target-groups/${uid}/results`,
    params
  )
}

/** POST /v1/probe-target-groups/{uid}/probe-now */
export function probeTargetGroupNow(
  uid: string
): Promise<{ results: ProbeResultItem[] }> {
  return post<{ results: ProbeResultItem[] }>(
    `/probe-target-groups/${uid}/probe-now`,
    { uid }
  )
}

/** POST /v1/probe-trial */
export function probeTrial(
  body: ProbeTrialParams
): Promise<ProbeTrialResult> {
  return post<ProbeTrialResult>("/probe-trial", body)
}

export type {
  CreateProbeTargetGroupParams,
  ListProbeTargetGroupsParams,
  ListProbeTargetGroupsResult,
  ListProbeTargetResultsParams,
  ListProbeTargetResultsResult,
  ProbeDriver,
  ProbeRequestConfig,
  ProbeResultItem,
  ProbeTargetGroupItem,
  ProbeTargetInput,
  ProbeTargetItem,
  ProbeTrialParams,
  ProbeTrialResult,
  UpdateProbeTargetGroupParams,
} from "./types"
