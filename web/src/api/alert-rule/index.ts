import { del, get, post, put } from "@/api/http"

import type {
  AlertRuleGroupItem,
  AlertRuleItem,
  ListAlertRuleGroupsParams,
  ListAlertRuleGroupsResult,
  ListAlertRulesParams,
  ListAlertRulesResult,
  SaveAlertRuleParams,
} from "./types"

/** GET /v1/alert-rule-groups */
export function listAlertRuleGroups(params: ListAlertRuleGroupsParams = {}): Promise<ListAlertRuleGroupsResult> {
  return get<ListAlertRuleGroupsResult>("/alert-rule-groups", params)
}

/** POST /v1/alert-rule-group */
export function createAlertRuleGroup(name: string): Promise<{ group: AlertRuleGroupItem }> {
  return post<{ group: AlertRuleGroupItem }>("/alert-rule-group", { name })
}

/** PUT /v1/alert-rule-group/{uid} */
export function updateAlertRuleGroup(uid: string, name: string): Promise<{ group: AlertRuleGroupItem }> {
  return put<{ group: AlertRuleGroupItem }>(`/alert-rule-group/${uid}`, { uid, name })
}

/** PUT /v1/alert-rule-group/{uid}/status */
export function updateAlertRuleGroupStatus(uid: string, status: string): Promise<{ group: AlertRuleGroupItem }> {
  return put<{ group: AlertRuleGroupItem }>(`/alert-rule-group/${uid}/status`, { uid, status })
}

/** DELETE /v1/alert-rule-group/{uid} */
export function deleteAlertRuleGroup(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-rule-group/${uid}`)
}

/** GET /v1/alert-rules */
export function listAlertRules(params: ListAlertRulesParams = {}): Promise<ListAlertRulesResult> {
  return get<ListAlertRulesResult>("/alert-rules", params)
}

/** GET /v1/alert-rule/{uid} */
export function getAlertRule(uid: string): Promise<{ rule: AlertRuleItem }> {
  return get<{ rule: AlertRuleItem }>(`/alert-rule/${uid}`)
}

/** POST /v1/alert-rule */
export function createAlertRule(body: SaveAlertRuleParams): Promise<{ rule: AlertRuleItem }> {
  return post<{ rule: AlertRuleItem }>("/alert-rule", body)
}

/** PUT /v1/alert-rule/{uid} */
export function updateAlertRule(body: SaveAlertRuleParams & { uid: string }): Promise<{ rule: AlertRuleItem }> {
  return put<{ rule: AlertRuleItem }>(`/alert-rule/${body.uid}`, body)
}

/** PUT /v1/alert-rule/{uid}/status */
export function updateAlertRuleStatus(uid: string, status: string): Promise<{ rule: AlertRuleItem }> {
  return put<{ rule: AlertRuleItem }>(`/alert-rule/${uid}/status`, { uid, status })
}

/** DELETE /v1/alert-rule/{uid} */
export function deleteAlertRule(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/alert-rule/${uid}`)
}

export type {
  AlertRuleGroupItem,
  AlertRuleItem,
  AlertRuleLevelInput,
  ListAlertRuleGroupsParams,
  ListAlertRulesParams,
  SaveAlertRuleParams,
} from "./types"
