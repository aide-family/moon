import { del, get, post, put } from "@/api/http"

import type {
  AlertEscalationPolicyItem,
  ListAlertEscalationPoliciesParams,
  ListAlertEscalationPoliciesResult,
  ListEventEscalationsResult,
  SaveAlertEscalationPolicyParams,
} from "./types"

/** GET /v1/alert-escalation-policies */
export function listAlertEscalationPolicies(
  params: ListAlertEscalationPoliciesParams = {}
): Promise<ListAlertEscalationPoliciesResult> {
  return get<ListAlertEscalationPoliciesResult>(
    "/alert-escalation-policies",
    params
  )
}

/** GET /v1/alert-escalation-policies/{uid} */
export function getAlertEscalationPolicy(
  uid: string
): Promise<{ policy: AlertEscalationPolicyItem }> {
  return get<{ policy: AlertEscalationPolicyItem }>(
    `/alert-escalation-policies/${uid}`
  )
}

/** POST /v1/alert-escalation-policies */
export function createAlertEscalationPolicy(
  body: SaveAlertEscalationPolicyParams
): Promise<{ policy: AlertEscalationPolicyItem }> {
  return post<{ policy: AlertEscalationPolicyItem }>(
    "/alert-escalation-policies",
    body
  )
}

/** PUT /v1/alert-escalation-policies/{uid}：步骤整体替换，按 uid 匹配已有步骤。 */
export function updateAlertEscalationPolicy(
  body: SaveAlertEscalationPolicyParams & { uid: string }
): Promise<{ policy: AlertEscalationPolicyItem }> {
  return put<{ policy: AlertEscalationPolicyItem }>(
    `/alert-escalation-policies/${body.uid}`,
    body
  )
}

/** PUT /v1/alert-escalation-policies/{uid}/status */
export function updateAlertEscalationPolicyStatus(
  uid: string,
  status: string
): Promise<{ policy: AlertEscalationPolicyItem }> {
  return put<{ policy: AlertEscalationPolicyItem }>(
    `/alert-escalation-policies/${uid}/status`,
    { uid, status }
  )
}

/** DELETE /v1/alert-escalation-policies/{uid} */
export function deleteAlertEscalationPolicy(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-escalation-policies/${uid}`)
}

/** GET /v1/event-escalations?eventUid=：权限 alert_realtime:read。 */
export function listEventEscalations(
  eventUid: string
): Promise<ListEventEscalationsResult> {
  return get<ListEventEscalationsResult>("/event-escalations", { eventUid })
}

export type {
  AlertEscalationPolicyItem,
  AlertEscalationPolicyStatus,
  AlertEscalationStepInput,
  AlertEscalationStepItem,
  EventEscalationItem,
  EventEscalationStatus,
  ListAlertEscalationPoliciesParams,
  ListAlertEscalationPoliciesResult,
  ListEventEscalationsResult,
  SaveAlertEscalationPolicyParams,
} from "./types"
